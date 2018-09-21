package gopl

import (
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-pipeline/util"
	"reflect"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

type InputFactory func() Input
type FilterFactory func() Filter
type OutputFactory func() Output
type HookFunc func()
type Component interface{}

// 根据组件类型，自动注册到对应的组件中
func (slf *GoPipeline) AutoRegister(component Component) {
	plgType := reflect.TypeOf(component)
	if plgType.Kind() != reflect.Ptr {
		log.Panic().Msg("AutoRegister: Needs a pointer to a component")
	}
	typeName := util.SimpleTypeName(plgType)
	log.Info().Msgf("AutoRegister: <%s>, type: <%s>", typeName, plgType)

	rawType := plgType.Elem()

	switch component.(type) {

	case Plugin:
		slf.RegisterPlugin(component.(Plugin))

	case Decoder:
		slf.RegisterDecoder(typeName, component.(Decoder))

	case Matcher:
		slf.RegisterMatcher(typeName, component.(Matcher))

	case Input:
		slf.RegisterInput(typeName, func() Input {
			return reflect.New(rawType).Interface().(Input)
		})

	case Filter:
		slf.RegisterFilter(typeName, func() Filter {
			return reflect.New(rawType).Interface().(Filter)
		})

	case Output:
		slf.RegisterOutput(typeName, func() Output {
			return reflect.New(rawType).Interface().(Output)
		})

	default:
		log.Panic().Msgf("AutoRegister: Unsupported component type, was: %s", typeName)
	}
}

// RegisterStartupHook 注册一个启动Hook。此Hook将会在程序启动初始化组件前调用
func (slf *GoPipeline) RegisterStartupHook(h HookFunc) {
	slf.startupHook.PushBack(h)
}

// RegisterShutdownHook 注册一个停止Hook。此Hook将会在程序停止所有组件后调用
func (slf *GoPipeline) RegisterShutdownHook(h HookFunc) {
	slf.shutdownHook.PushBack(h)
}

func (slf *GoPipeline) RegisterTerminateHook(h HookFunc) {
	slf.RegisterShutdownHook(h)
}

// 注册一个插件
func (slf *GoPipeline) RegisterPlugin(plg Plugin) {
	slf.plugins.PushBack(plg)
}

//// 三大组件类型注册

func (slf *GoPipeline) RegisterInput(typeName string, factory InputFactory) {
	slf.factoryInputs[typeName] = factory
}

func (slf *GoPipeline) RegisterFilter(pluginName string, factory FilterFactory) {
	slf.factoryFilters[pluginName] = factory
}

func (slf *GoPipeline) RegisterOutput(pluginName string, factory OutputFactory) {
	slf.factoryOutputs[pluginName] = factory
}

func (slf *GoPipeline) RegisterDecoder(typeName string, decoder Decoder) {
	slf.decoders[typeName] = decoder
}

func (slf *GoPipeline) RegisterDecoderOf(decoder Decoder) {
	slf.RegisterDecoder(util.SimpleClassName(decoder), decoder)
}

func (slf *GoPipeline) RegisterMatcher(typeName string, matcher Matcher) {
	slf.matchers[typeName] = matcher
}

func (slf *GoPipeline) RegisterMatcherOf(matcher Matcher) {
	slf.RegisterMatcher(util.SimpleClassName(matcher), matcher)
}
