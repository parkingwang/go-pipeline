package abc

import (
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog/log"
	"time"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
// 定时调度器抽象实现
//

type AbcScheduler struct {
	AbcShutdown
	Interval time.Duration
}

func (slf *AbcScheduler) Init(args conf.Map) {
	slf.AbcShutdown.Init()
	slf.Interval = args.MustDuration("interval")
	if slf.Interval <= 0 {
		log.Panic().Msgf("Invalid interval in args: %s", args)
	}
}

func (slf *AbcScheduler) OnTick(tick func(c time.Time)) {
	defer slf.SetTerminated()

	ticker := time.NewTicker(slf.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-slf.ShutdownChan():
			return

		case c := <-ticker.C:
			tick(c)
		}
	}
}
