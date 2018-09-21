package gopl

import (
	"bytes"
	"io"
	"strings"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
// 解码器

// 解码器用于将Input插件获取的数据解码成DataFrame对象。
// 每个Input插件的解码器由配置文件指定（默认为ObtainDecoder），Input插件在获取数据后，可以使用解码器来对数据解码，生成DataFrame对象。
// 在使用上，解码器并非必须的，如果非通用Input，可以忽略解码器，直接生成DataFrame来发送。
type Decoder interface {
	// 解码数据，返回DataFrame对象
	Decode(data interface{}) (*DataFrame, error)
}

////

const TypeNameJSONDecoder = "JSONDecoder"

// JSON字节解码器
type JSONDecoder struct {
	Decoder
}

func (*JSONDecoder) Decode(data interface{}) (*DataFrame, error) {
	pack := ObtainDataFrame()
	switch v := data.(type) {
	case []byte:
		pack.SetBody(bytes.NewBuffer(v))

	case string:
		pack.SetBody(strings.NewReader(v))

	case io.Reader:
		pack.SetBody(v)

	default:
		if buf, err := MarshalJSON(data); nil != err {
			return nil, err
		} else {
			pack.SetBody(bytes.NewBuffer(buf))
		}
	}
	return pack, nil
}
