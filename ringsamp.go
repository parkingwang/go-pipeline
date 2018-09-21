package gopl

import (
	"strconv"
	"sync"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
// 环形数值采样
//

type IntSamples struct {
	max    int
	data   []int64
	cursor int
	mu     *sync.RWMutex
}

func NewIntSamples(maxSize int) *IntSamples {
	if maxSize < 0 {
		panic("Invalid size: " + strconv.Itoa(maxSize))
	}

	return &IntSamples{
		max:    maxSize,
		data:   make([]int64, maxSize),
		cursor: -1,
		mu:     new(sync.RWMutex),
	}
}

func (slf *IntSamples) AddSample(val int64) {
	slf.mu.Lock()
	defer slf.mu.Unlock()
	slf.cursor = (slf.cursor + 1) % slf.max
	slf.data[slf.cursor] = val
}

func (slf *IntSamples) Avg() int64 {
	slf.mu.RLock()
	defer slf.mu.RUnlock()
	var sum int64
	for _, iv := range slf.data {
		sum += iv
	}
	return sum / int64(slf.max)
}
