package common

import (
	"bufio"
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-pipeline"
	"github.com/yoojia/go-pipeline/abc"
	"github.com/yoojia/go-pipeline/splitter"
	"os"
	"time"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

type GoPLFilePollingInput struct {
	gopl.AbcSlot
	abc.AbcShutdown

	filePath string
	interval time.Duration
}

func (slf *GoPLFilePollingInput) Init(args conf.Map) {
	slf.AbcSlot.Init(args)
	slf.AbcShutdown.Init()
	slf.interval = gopl.DurationValue(args.MustString("interval"))
	slf.filePath = args.MustString("file_path")
	if "" == slf.filePath {
		slf.TagLog(log.Panic).Msgf("Param <file_path> is required for Input:[%s]", slf.GetName())
	}
}

func (slf *GoPLFilePollingInput) Input(deliverer gopl.Deliverer, decoder gopl.Decoder) {
	defer slf.SetTerminated()

	bytesSplitter := new(splitter.BytesSplitter)

	ticker := time.NewTicker(slf.interval)
	defer ticker.Stop()

	vv := gopl.Debugs().VeryVerbose

	received := func(bytes []byte) {
		pack, err := decoder.Decode(bytes)
		if nil != err {
			slf.TagLog(log.Error).Err(err).Msgf("Error when decode content from file: %s", slf.filePath)
		} else {
			deliverer.Deliver(pack)
		}
	}

	polling := func() {
		file, ferr := os.Open(slf.filePath)
		if nil != ferr {
			slf.TagLog(log.Error).Err(ferr).Msgf("Error when open file: %s", slf.filePath)
			return
		}
		defer file.Close()

		if vv {
			slf.TagLog(log.Debug).Msgf("Reading file: %s", slf.filePath)
		}

		serr := bytesSplitter.Split(bufio.NewReader(file), &splitter.SplitterEvents{
			OnReceived: received,
		})
		if nil != serr {
			slf.TagLog(log.Error).Err(serr).Msgf("Split FAILED, file: %s", slf.filePath)
		}
	}

	for {
		select {
		case <-slf.ShutdownChan():
			return

		case <-ticker.C:
			polling()
		}
	}
}
