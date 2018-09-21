package util

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func MaxInt(a int, b int) int {
	return int(MaxInt32(int32(a), int32(b)))
}

func MinInt(a int, b int) int {
	return int(MinInt32(int32(a), int32(b)))
}

func MaxInt32(a int32, b int32) int32 {
	if a > b {
		return a
	} else {
		return b
	}
}

func MinInt32(a int32, b int32) int32 {
	if a < b {
		return a
	} else {
		return b
	}
}
