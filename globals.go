package gopl

import "time"

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

var (
	_gStartupTimestamp_ = time.Now()
)

////

// StartupTime 返回程序启动的时间
func StartupTime() time.Time {
	return _gStartupTimestamp_
}

// Uptime 返回到当前为止程序的运行时间
func Uptime() time.Duration {
	return time.Since(_gStartupTimestamp_)
}
