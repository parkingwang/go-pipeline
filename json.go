package gopl

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//
import (
	"github.com/json-iterator/go"
	"github.com/pkg/errors"
)

var (
	gJsonParser = jsoniter.ConfigFastest
)

// MarshalJSON 将指定值/对象序列化成 byte 数组
func MarshalJSON(value interface{}) ([]byte, error) {
	if out, err := gJsonParser.Marshal(value); nil != err {
		return out, errors.WithMessage(err, "marshal json")
	} else {
		return out, nil
	}
}

// UnmarshalJSON 将指定 byte 数组，反序列化指定对象和结构。
func UnmarshalJSON(bytes []byte, out interface{}) error {
	if err := gJsonParser.Unmarshal(bytes, out); nil != err {
		return errors.WithMessage(err, "unmarshal json")
	} else {
		return nil
	}
}
