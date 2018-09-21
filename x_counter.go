package gopl

import "sync/atomic"

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 消息数据统计
type FioCounter struct {
	InCount  uint64
	OutCount uint64
	FilCount uint64
}

func (slf *FioCounter) Inbounds() uint64 {
	return atomic.LoadUint64(&slf.InCount)
}

func (slf *FioCounter) Outbounds() uint64 {
	return atomic.LoadUint64(&slf.OutCount)
}

func (slf *FioCounter) Filtered() uint64 {
	return atomic.LoadUint64(&slf.FilCount)
}

var gFioCounter = &FioCounter{}

// 重置统计数据
func resetFioCounter() {
	atomic.StoreUint64(&gFioCounter.InCount, 0)
	atomic.StoreUint64(&gFioCounter.FilCount, 0)
	atomic.StoreUint64(&gFioCounter.OutCount, 0)
}

func increaseInbound() {
	atomic.AddUint64(&gFioCounter.InCount, 1)
}

func increaseFilter() {
	atomic.AddUint64(&gFioCounter.FilCount, 1)
}

func increaseOutbounds() {
	atomic.AddUint64(&gFioCounter.OutCount, 1)
}

func GetFioCounter() *FioCounter {
	return gFioCounter
}
