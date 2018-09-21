package common

import (
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-jsonx"
	"github.com/yoojia/go-pipeline"
	"github.com/yoojia/go-pipeline/abc"
	"time"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

type TplArgs struct {
	DataUri string
}

type GoPLDeliverCountInput struct {
	gopl.AbcSlot
	abc.AbcScheduler
}

func (slf *GoPLDeliverCountInput) Init(args conf.Map) {
	slf.AbcSlot.Init(args)
	slf.AbcScheduler.Init(args)
}

func (slf *GoPLDeliverCountInput) Input(deliverer gopl.Deliverer, decoder gopl.Decoder) {
	inboundCount := uint64(0)
	filtersCount := uint64(0)
	outputsCount := uint64(0)

	slf.OnTick(func(c time.Time) {
		stats := gopl.GetFioCounter()
		ic := stats.Inbounds()
		fc := stats.Filtered()
		oc := stats.Outbounds()

		// 统计每个周期的消息处理量
		json := jsonx.NewFatJSON()
		json.Field("type", "deliver.count")
		json.Field("time", time.Now().Format("15:04:05"))
		json.FieldNotEscapeValue("inbound", ic-inboundCount)
		json.FieldNotEscapeValue("filter", fc-filtersCount)
		json.FieldNotEscapeValue("outbound", oc-outputsCount)

		avg := gopl.TakeDataFramesAvgSamples()
		json.FieldNotEscapeValue("avg.inbound", avg.InboundsAvg)
		json.FieldNotEscapeValue("avg.filter", avg.FilteredAvg)
		json.FieldNotEscapeValue("avg.outbound", avg.OutboundsAvg)

		inboundCount = ic
		filtersCount = fc
		outputsCount = oc

		bytes := json.Bytes()
		if pack, err := decoder.Decode(bytes); nil != err {
			slf.TagLog(log.Error).Err(err).RawJSON("json", bytes).Msg("Failed to decode from text")
		} else {
			deliverer.Deliver(pack)
		}
	})
}
