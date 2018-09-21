package common

import (
	"bufio"
	"bytes"
	"github.com/pkg/errors"
	"github.com/yoojia/go-jsonx"
	"github.com/yoojia/go-pipeline"
	"reflect"
	"strings"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

type GoPLProcMemInfoDecoder struct {
	gopl.Decoder
}

/**
MemTotal:       12161600 kB
MemFree:         1617820 kB
MemAvailable:    7005452 kB
Buffers:         1508092 kB
Cached:          3488376 kB
*/
func (slf *GoPLProcMemInfoDecoder) Decode(data interface{}) (*gopl.DataFrame, error) {
	txt := ""
	switch v := data.(type) {
	case []byte:
		txt = string(v)

	case string:
		txt = v

	default:
		return nil, errors.Errorf("Decode data is not bytes or string, was: %s", reflect.TypeOf(data))
	}
	scanner := bufio.NewScanner(strings.NewReader(txt))
	buf := jsonx.NewFatJSON()
	for scanner.Scan() {
		fields := strings.Split(string(scanner.Bytes()), ":")
		if 2 == len(fields) {
			name := strings.TrimSpace(fields[0])
			value := strings.TrimSpace(fields[1])
			buf.Field(name, value)
		}
	}
	if err := scanner.Err(); nil != err {
		return nil, err
	}
	pack := gopl.ObtainDataFrame()
	pack.SetBody(bytes.NewBuffer(buf.Bytes()))
	return pack, nil
}
