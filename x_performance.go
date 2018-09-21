package gopl

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 统计消息处理性能的最大消息数量
const MaxSamplesCountPerformanceAvg = 1024

// 消息处理性能
type DataFrameSamples struct {
	InboundSamples  *IntSamples
	OutboundSamples *IntSamples
	FilteredSamples *IntSamples
}

var gPerformanceSamples = DataFrameSamples{
	InboundSamples:  NewIntSamples(MaxSamplesCountPerformanceAvg),
	OutboundSamples: NewIntSamples(MaxSamplesCountPerformanceAvg),
	FilteredSamples: NewIntSamples(MaxSamplesCountPerformanceAvg),
}

func sampleInbound(du int64) {
	gPerformanceSamples.InboundSamples.AddSample(du)
}

func sampleFilter(du int64) {
	gPerformanceSamples.FilteredSamples.AddSample(du)
}

func sampleOutbounds(du int64) {
	gPerformanceSamples.OutboundSamples.AddSample(du)
}

type AvgSamples struct {
	InboundsAvg  int64
	OutboundsAvg int64
	FilteredAvg  int64
}

func TakeDataFramesAvgSamples() AvgSamples {
	return AvgSamples{
		InboundsAvg:  gPerformanceSamples.InboundSamples.Avg(),
		OutboundsAvg: gPerformanceSamples.OutboundSamples.Avg(),
		FilteredAvg:  gPerformanceSamples.FilteredSamples.Avg(),
	}
}
