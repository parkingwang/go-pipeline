package gopl

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"time"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
// 消息输入接口 Input
//

const FieldNameDataFrameHeaders = "DataFrameHeaders" // Input注入消息Header时使用的配置字段名

// 消息投递接口
type Deliverer interface {
	// 发送消息
	Deliver(msg *DataFrame)
}

// 消息输入接口
type Input interface {
	VirtualSlot
	// 生产和通过Deliverer投递消息。
	// Decoder是为Input配置的数据解码接口。
	Input(deliverer Deliverer, decoder Decoder)
}

////

type inputRunner struct {
	input     Input
	decoder   Decoder
	config    *ComponentConfig
	configKey string
}

func newInputRunner(input Input, decoder Decoder, config *ComponentConfig, configKey string) *inputRunner {
	return &inputRunner{
		input:     input,
		decoder:   decoder,
		config:    config,
		configKey: configKey,
	}
}

func (slf *inputRunner) init() {
	pluginName := slf.configKey
	slf.input.SetName(pluginName)

	// Init
	log.Info().Msgf("Init Input: <%s>, decoder: <%T>", pluginName, slf.decoder)

	slf.input.Init(slf.config.InitArgs)
}

func (slf *inputRunner) start(deliverer Deliverer) {
	pluginName := slf.input.GetName()
	headerValue := slf.config.InitArgs.MustMap(FieldNameDataFrameHeaders)
	if 0 < len(headerValue) {
		log.Info().Msgf("Init Input: <%s> with headers: %s", pluginName, headerValue)
	}
	headers := make(map[string]string)
	for k, v := range headerValue {
		headers[k] = fmt.Sprintf("%v", v)
	}
	log.Info().Msgf("Start Input: <%s>", pluginName)
	proxy := &delivererProxy{
		realDeliverer: deliverer,
		signer:        pluginName,
		injectHeaders: headers,
		injectTopic:   slf.config.Topic,
	}
	slf.input.Input(proxy, slf.decoder)
}

////

type delivererProxy struct {
	realDeliverer Deliverer
	signer        string
	injectHeaders Headers
	injectTopic   string
}

// 发送消息
func (slf *delivererProxy) Deliver(pack *DataFrame) {
	ts := time.Now()
	pack.SetHeader("Origin", slf.signer)
	pack.addTrace(slf.signer, ts.UnixNano())
	pack.SetHeaders(slf.injectHeaders)
	pack.setTopic(slf.injectTopic)
	slf.realDeliverer.Deliver(pack)

	// Counting and Samples
	go func() {
		increaseInbound()
		sampleInbound(time.Now().Sub(ts).Nanoseconds())
	}()
}
