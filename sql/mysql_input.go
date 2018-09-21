package sql

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-pipeline"
	"github.com/yoojia/go-pipeline/abc"
	"strconv"
	"strings"
	"time"
)

type GoPLMySQLQueryInput struct {
	gopl.AbcSlot
	abc.AbcShutdown
	interval time.Duration

	dataSource       string
	querySQL         string
	queryLimitStart  time.Time     // 起始时间
	queryLimitEnd    time.Time     // 结束时间
	queryTimeSection time.Duration // 查询时间距离
	queryTimestamp   time.Time     // 查询当前时间
}

func (slf *GoPLMySQLQueryInput) Init(args conf.Map) {
	slf.AbcSlot.Init(args)
	slf.AbcShutdown.Init()
	slf.interval = args.MustDuration("interval")
	slf.dataSource = args.MustString("db_data_source")
	slf.querySQL = args.MustString("db_query_sql")

	start := args.MustString("db_query_start")
	if "" != start {
		if ts, err := time.Parse("2006-01-02 15:04:05", start); nil != err {
			slf.TagLog(log.Panic).Err(err).Msgf("Invalid <query_start>: ", start)
		} else {
			slf.queryLimitStart = ts
		}
	} else {
		slf.queryLimitStart = time.Unix(0, 0)
	}

	end := args.MustString("db_query_end")
	if "" != end {
		// 配置End时，必须配置Start
		if slf.queryLimitStart.IsZero() {
			slf.TagLog(log.Panic).Msg("Require <db_query_start> when set <db_query_end>")
		}

		if rangeEnd, err := time.Parse("2006-01-02 15:04:05", end); nil != err {
			slf.TagLog(log.Panic).Err(err).Msgf("Invalid <db_query_end>: ", end)
		} else {
			slf.queryLimitEnd = rangeEnd
		}
	} else {
		slf.queryLimitEnd = time.Unix(0, 0)
	}

	slf.queryTimeSection = args.GetDurationOrDefault("db_query_section", time.Hour*24)

	slf.queryTimestamp = slf.queryLimitStart

	slf.TagLog(log.Info).Msgf("Query range: [%s - %s], query.du: %s", slf.queryLimitStart, slf.queryLimitEnd, slf.queryTimeSection)
}

func (slf *GoPLMySQLQueryInput) Input(deliverer gopl.Deliverer, decoder gopl.Decoder) {
	defer slf.SetTerminated()

	queryTicker := time.NewTicker(slf.interval)
	defer queryTicker.Stop()

	db, err := sqlx.Connect("mysql", slf.dataSource)
	if nil != err {
		slf.TagLog(log.Panic).Err(err).Msgf("Database connection failed: %s", slf.dataSource)
	} else {
		slf.TagLog(log.Info).Msgf("Database connected")
	}
	defer db.Close()

	for {
		select {
		case <-slf.ShutdownChan():
			return

		case <-queryTicker.C:
			if err := slf.queryInRange(db, deliverer, decoder); nil != err {
				slf.TagLog(log.Error).Err(err).Msg(err.Error())
				// SQL错误，检查DB连接
				if e := db.Ping(); nil != e {
					slf.TagLog(log.Info).Msgf("Database reconnecting")
					if ndb, err := sqlx.Connect("mysql", slf.dataSource); nil != err {
						slf.TagLog(log.Error).Err(err).Msg(err.Error())
					} else {
						slf.TagLog(log.Info).Msgf("Database reconnected")
						db = ndb
					}
				}
			} else {
				slf.TagLog(log.Info).Msg("Finish query task")
			}
		}
	}
}

func (slf *GoPLMySQLQueryInput) queryInRange(db *sqlx.DB, deliverer gopl.Deliverer, decoder gopl.Decoder) error {
	query, next := slf.nextQuerySQL()
	slf.TagLog(log.Info).Str("sql", query).Msg("Executing SQL")
	if !next {
		return nil
	}

	rows, err := db.Queryx(query)
	if nil != err {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		values := make(map[string]interface{})
		if err := rows.MapScan(values); nil != err {
			return err
		} else {
			if msg, err := decoder.Decode(values); nil != err {
				return err
			} else {
				deliverer.Deliver(msg)
			}
		}
	}

	// 更新查询条件
	if next {
		slf.queryTimestamp = slf.queryTimestamp.Add(slf.queryTimeSection)
	}

	return nil
}

func (slf *GoPLMySQLQueryInput) nextQuerySQL() (string, bool) {
	if strings.Contains(slf.querySQL, "$start") && strings.Contains(slf.querySQL, "$end") {
		// 检查是否超过时间范围
		if !slf.queryLimitEnd.IsZero() && slf.queryTimestamp.After(slf.queryLimitEnd) {
			return "", false
		}
		start := slf.queryTimestamp.Unix()
		end := slf.queryTimestamp.Add(slf.queryTimeSection).Unix()

		out := strings.Replace(slf.querySQL, "$start", strconv.FormatInt(start, 10), -1)
		out = strings.Replace(out, "$end", strconv.FormatInt(end, 10), -1)
		return out, true
	} else {
		return slf.querySQL, true
	}
}
