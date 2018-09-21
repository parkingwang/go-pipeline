package gopl

import (
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog/log"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//
const (
	componentUnknown = iota
	componentInput
	componentFilter
	componentOutput
)

// ifCoreComponentConfig 返回此配置是否为核心组件的配置
func ifCoreComponentConfig(configName string, config conf.Map) (int, string, bool) {
	// 1. 先直接查找其配置名
	if plgType, typeName, ok := ifHasFactoryFuncOfType(configName); ok {
		return plgType, typeName, true
	}
	// 2. 再查找其plugin字段
	if pluginName, hit := config[componentTypeFieldName]; !hit {
		return componentUnknown, "", false
	} else {
		return ifHasFactoryFuncOfType(pluginName.(string))
	}
}

// 是否为已经注册的插件类型名
func ifHasFactoryFuncOfType(typeName string) (int, string, bool) {
	if _, hit := SharedRouter().factoryInputs[typeName]; hit {
		return componentInput, typeName, true
	}

	if _, hit := SharedRouter().factoryFilters[typeName]; hit {
		return componentFilter, typeName, true
	}

	if _, hit := SharedRouter().factoryOutputs[typeName]; hit {
		return componentOutput, typeName, true
	}
	return componentUnknown, "", false
}

func findNonNilDecoder(input Input, config *ComponentConfig, pluginName string) Decoder {
	const notFound = "Decoder: <%s> sets but not found, for Output: <%s>"

	if 0 < len(config.DecoderName) {
		if decoder := SharedRouter().decoders[config.DecoderName]; decoder != nil {
			return decoder
		} else {
			log.Panic().Msgf(notFound, config.DecoderName, pluginName)
		}
	}

	// then, defaults
	if defaults, ok := input.(DefaultsDecoder); ok {
		defaultDecoder := defaults.UseDefaultDecoder()
		switch defaultDecoder.(type) {
		case string:
			typeName := defaultDecoder.(string)
			defDecoder := SharedRouter().decoders[typeName]
			if nil == defDecoder {
				log.Panic().Msgf(notFound, typeName, pluginName)
			}
			return defDecoder

		case Decoder:
			return defaultDecoder.(Decoder)

		default:
			log.Panic().Msgf("Only accepts [<string>, <Decoder>] for Input:<%s>.Decoder registry", pluginName)
		}
	}

	log.Info().Msgf("Decoder is NOT SET for Input: <%s>, use default.", pluginName)
	// 默认Decoder为JSONDecoder
	return SharedRouter().decoders[TypeNameJSONDecoder]
}

func findNonNilMatcher(plugin VirtualSlot, conf *ComponentConfig) Matcher {
	// 首先检查Topic字段是否配置
	if "" != conf.Topic {
		// Match any messages
		if "*" == conf.Topic {
			return new(AnyMatcher)
		}
		// By URL
		matcher, err := NewDefaultURLMatcher(conf.Topic)
		if nil != err {
			log.Panic().Err(err).Msgf("Parse matcher from topic FAILED, topic: ", conf.Topic)
		}
		return matcher
	}

	// 其它默认接口实现
	if defaults, ok := plugin.(DefaultsMatcher); ok {
		plgName := plugin.GetName()

		defaultMatcher := defaults.UseDefaultMatcher()
		switch defaultMatcher.(type) {
		case string:
			typeName := defaultMatcher.(string)
			matcher := SharedRouter().matchers[typeName]
			if nil == matcher {
				log.Panic().Msgf("Matcher: <%s> sets but NOT FOUND, for Output: <%s>", typeName, plgName)
			}
			return matcher

		case Matcher:
			return defaultMatcher.(Matcher)

		default:
			log.Panic().Msgf("Only accepts [<string>, <Matcher>]", plgName)
		}
	}

	return new(NoneMatcher)
}
