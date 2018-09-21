package gopl

import (
	"github.com/rs/zerolog/log"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
// 消息输出接口
//

type Output interface {
	VirtualSlot
	// 处理消息
	Output(pack *DataFrame)
}

type outputRunner struct {
	output    Output
	matcher   Matcher
	config    *ComponentConfig
	configKey string
}

func newOutputRunner(output Output, matcher Matcher, config *ComponentConfig, configKey string) *outputRunner {
	return &outputRunner{
		output:    output,
		matcher:   matcher,
		config:    config,
		configKey: configKey,
	}
}

func (slf *outputRunner) init() {
	pluginName := slf.configKey
	slf.output.SetName(pluginName)

	log.Info().Msgf("Init Output: <%s>, matcher: <%T>", pluginName, slf.matcher)
	go slf.output.Init(slf.config.InitArgs)
}

func (slf *outputRunner) runOutput(pack *DataFrame) {
	slf.output.Output(pack)
}

func (slf *outputRunner) checkAccept(pack *DataFrame) bool {
	return slf.matcher.Match(pack)
}
