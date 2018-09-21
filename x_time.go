package gopl

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//
import (
	"github.com/rs/zerolog/log"
	"time"
)

func DurationValue(du string) time.Duration {
	return DurationOrDefault(du, 0)
}

func DurationOrDefault(str string, defaultT time.Duration) time.Duration {
	if "" == str {
		return defaultT
	}
	if duration, err := time.ParseDuration(str); nil != err {
		log.Error().Err(err).Msg(`Invalid period value, such as "200ms", "3s", use default.`)
		return defaultT
	} else {
		return duration
	}
}
