package gopl

import (
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

type AbcSlot struct {
	slotName string
	args     conf.Map
}

func (slf *AbcSlot) Args() conf.Map {
	return slf.args
}

func (slf *AbcSlot) Init(args conf.Map) {
	slf.args = args
}

// GetName 返回插件名称。
func (slf *AbcSlot) GetName() string {
	return slf.slotName
}

// SetName 由框架内部在启动时调用，设置插件的名称。
func (slf *AbcSlot) SetName(name string) {
	slf.slotName = name
}

func (slf *AbcSlot) TagLog(f func() *zerolog.Event) *zerolog.Event {
	return f().Str("tag", slf.GetName())
}
