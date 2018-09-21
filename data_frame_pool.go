package gopl

import "sync"

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 全局对象缓存池
var gDataFramePool = &sync.Pool{
	New: func() interface{} {
		return NewDataFrame()
	},
}

// 从对象池中申请一个消息对象
func ObtainDataFrame() *DataFrame {
	return gDataFramePool.Get().(*DataFrame)
}

// 创建一个消息对象，不从全局对象池中获取。
func NewDataFrame() *DataFrame {
	out := &DataFrame{
		MultiReader: NewMultiReader(),
	}
	out.headers = make(Headers)
	out.traces = make([]*Trace, 6)
	out.topic = ""
	return out
}

// 将消息对象释放，重置对象数据，并放回对象池中。
func releaseDataFrame(df *DataFrame) {
	df.Close()
	for k := range df.headers {
		delete(df.headers, k)
	}
	for i := range df.traces {
		df.traces[i] = nil
	}
	df.topic = ""
	gDataFramePool.Put(df)
}
