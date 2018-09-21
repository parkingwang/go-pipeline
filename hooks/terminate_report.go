package hooks

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//
import (
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-pipeline"
	"github.com/yoojia/go-unit"
	"time"
)

func TPSReporterTerminateHook() {
	du := gopl.Uptime()
	sec := du.Seconds()

	tps := unit.NewGoUnit(1, "").
		Next(1000, "K").
		Next(10, "W").
		SetUseBestMatchUnit(true)

	stats := gopl.GetFioCounter()
	inCNT := float64(stats.Inbounds())
	fiCNT := float64(stats.Filtered())
	otCNT := float64(stats.Outbounds())

	log.Info().Msgf("Uptime[INBOUNDS] CNT: %s, TPS: %s", tps.Format(inCNT), tps.Format(inCNT/sec))
	log.Info().Msgf("Uptime[FILTERED] CNT: %s, TPS: %s", tps.Format(fiCNT), tps.Format(fiCNT/sec))
	log.Info().Msgf("Uptime[OUTBOUND] CNT: %s, TPS: %s", tps.Format(otCNT), tps.Format(otCNT/sec))

	log.Info().Msgf("Started at: %s", gopl.StartupTime())
	log.Info().Msgf("Stopped at: %s", time.Now())
	log.Info().Msgf("Run duration : %s", du)
}
