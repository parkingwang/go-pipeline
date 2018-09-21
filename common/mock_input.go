package common

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//
import (
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-jsonx"
	"github.com/yoojia/go-pipeline"
	"github.com/yoojia/go-pipeline/abc"
	"time"
)

type GoPLMockInput struct {
	gopl.AbcSlot
	abc.AbcShutdown
	interval time.Duration
}

func (slf *GoPLMockInput) Init(args conf.Map) {
	slf.AbcSlot.Init(args)
	slf.AbcShutdown.Init()
	slf.interval = args.MustDuration("interval")
}

func (slf *GoPLMockInput) Input(deliverer gopl.Deliverer, decoder gopl.Decoder) {
	defer slf.SetTerminated()

	<-time.After(time.Millisecond)

	if 0 < slf.interval {
		slf.TagLog(log.Debug).Msgf("MockInput started with INTERVAL mode. interval: %v", slf.interval)
		slf.startTicker(deliverer, decoder)
	} else {
		slf.TagLog(log.Debug).Msg("MockInput started with INDEFINITELY mode")
		for {
			select {
			case <-slf.ShutdownChan():
				return

			default:
				deliverer.Deliver(newDataFrame(decoder))
			}
		}
	}
}

func (slf *GoPLMockInput) startTicker(deliverer gopl.Deliverer, decoder gopl.Decoder) {
	ticker := time.NewTicker(slf.interval)
	defer ticker.Stop()

	for {
		select {
		case <-slf.ShutdownChan():
			return

		case <-ticker.C:
			deliverer.Deliver(newDataFrame(decoder))

		}
	}
}

func newDataFrame(decoder gopl.Decoder) *gopl.DataFrame {
	return DecodeStringDataFrame(`
{
	"hook":
		{ 
			"type":"App", 
			"id":2008, 
			"active":true, 
			"events": [
				"pull_request", 
				"commit"
			], 
			"app_id":9991
		}
}`, decoder)
}

func DecodeStringDataFrame(txt string, decoder gopl.Decoder) *gopl.DataFrame {
	if pack, err := decoder.Decode(jsonx.CompressJSONBytes([]byte(txt))); nil != err {
		log.Error().Err(err).Str("text", txt).Msg("Decode mock text failed")
		return nil
	} else {
		return pack
	}
}
