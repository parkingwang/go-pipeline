package gopl

import (
	"fmt"
	"testing"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func TestNewRingArray(t *testing.T) {
	ra := NewIntSamples(10)
	for i := 0; i < 10; i++ {
		ra.AddSample(int64(i))
	}
	fmt.Println(ra.data)
	ra.AddSample(int64(99))
	ra.AddSample(int64(98))
	ra.AddSample(int64(97))
	fmt.Println(ra.data)
}

func BenchmarkIntSamples_AddSample(b *testing.B) {
	ra := NewIntSamples(b.N)
	for i := 0; i < b.N; i++ {
		ra.AddSample(int64(i))
	}
}
