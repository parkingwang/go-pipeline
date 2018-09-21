package gopl

import (
	"github.com/rs/zerolog/log"
	"time"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
// 消息处理接口

type Filter interface {
	VirtualSlot

	// 处理消息，并返回结果
	Filter(pack *DataFrame) *DataFrame
}

type NewFilterFactory func() Filter

type filterRunner struct {
	filter    Filter
	matcher   Matcher
	config    *ComponentConfig
	configKey string
}

func newFilterRunner(filter Filter, matcher Matcher, config *ComponentConfig, configKey string) *filterRunner {
	return &filterRunner{
		filter:    filter,
		matcher:   matcher,
		config:    config,
		configKey: configKey,
	}
}

func (slf *filterRunner) init(deliverer Deliverer) {
	pluginName := slf.configKey
	slf.filter.SetName(pluginName)
	// check deliverer supports
	if need, ok := slf.filter.(NeedDeliverer); ok {
		need.SetDeliverer(deliverer)
	}
	slf.filter.Init(slf.config.InitArgs)
	log.Info().Msgf("Init Filter: <%s>, matcher: <%T>", pluginName, slf.matcher)
}

func (slf *filterRunner) checkAccept(pack *DataFrame) bool {
	return slf.matcher.Match(pack)
}

func (slf *filterRunner) runFilter(ts time.Time, pack *DataFrame) *DataFrame {
	name := slf.filter.GetName()
	pack.addTrace(name, ts.UnixNano())
	// 处理并返回结果
	ret := slf.filter.Filter(pack)
	// 返回新结果时，复制Msg的基础参数
	if nil != ret && pack != ret {
		ret.SetHeader("Origin", name)
		ret.addTrace(name, ts.UnixNano())
		for k, v := range pack.headers {
			if _, hit := ret.headers[k]; !hit {
				ret.SetHeader(k, v)
			}
		}
		ret.setTopic(pack.topic)
	}
	return ret
}
