package gopl

import "github.com/parkingwang/go-conf"

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

type (
	// VirtualSlot 插槽是具有唯一命名的组件。可以通过命名来获取。
	VirtualSlot interface {
		Init(args conf.Map)
		SetName(name string)
		GetName() string
	}

	// Plugin 无命名组件，作为外挂组件存在。
	Plugin interface {
		Init()
		Startup()
		Shutdown()
	}
)

type (
	// Filter组件根据实现，是否需要启用Deliverer接口。
	// 如果启用，会在初始化时自动设置Deliverer接口。
	// 注意：此接口仅在Filter中使用。通过Deliverer，可以向消息处理流插入一条新的处理消息。
	NeedDeliverer interface {
		SetDeliverer(deliverer Deliverer)
		GetDeliverer() Deliverer
	}

	// 组件根据实现，是否支持Shutdown接口。
	// 如果启用，在程序关闭时会调用Shutdown接口。
	NeedShutdown interface {
		Shutdown()
	}
)

type (
	// Input组件内部实现，通过此接口自行提供 Decoder 默认实现，不需要从Config中配置。
	DefaultsDecoder interface {
		// 返回 Decoder 或者String类型名称
		UseDefaultDecoder() interface{}
	}

	// Input/Filter组件内部实现，通过此接口自行提供 Matcher 默认实现，不需要从Config中配置。
	DefaultsMatcher interface {
		// 返回 Matcher 或者String类型名称
		UseDefaultMatcher() interface{}
	}
)
