package splitter

import (
	"github.com/rs/zerolog/log"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 设置Splitter接口Setter、Getter默认实现
type AbcFeatureSplitter struct {
	splitter Splitter
}

// SetSplitter 由框架内部在初始化时设置Splitter
func (slf *AbcFeatureSplitter) SetSplitter(splitter Splitter) {
	slf.splitter = splitter
}

// GetSplitter 返回初始化时设置的Splitter
func (slf *AbcFeatureSplitter) GetSplitter() Splitter {
	return slf.splitter
}

// GetCheckedSplitter 检查并返回初始化时设置的Splitter。
// 如果Splitter没有设置，将会终止程序运行来报告错误。
func (slf *AbcFeatureSplitter) GetCheckedSplitter(messageIfNil string) Splitter {
	sp := slf.GetSplitter()
	if nil == sp {
		log.Fatal().Msg(messageIfNil)
	}
	return sp
}
