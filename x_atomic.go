package gopl

import "sync/atomic"

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

type AtomicBoolean struct {
	flag uint32
}

func NewAtomicBoolean() *AtomicBoolean {
	return &AtomicBoolean{
		flag: uint32(0),
	}
}

func (slf *AtomicBoolean) Get() bool {
	return 1 == atomic.LoadUint32(&slf.flag)
}

func (slf *AtomicBoolean) Set(flag bool) {
	val := uint32(0)
	if flag {
		val = 1
	} else {
		val = 0
	}
	atomic.StoreUint32(&slf.flag, val)
}
