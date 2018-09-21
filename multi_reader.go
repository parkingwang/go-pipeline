package gopl

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

type MultiReader struct {
	bodyLength  int64
	bodyRaw     io.ReadCloser
	bodyGetFunc func() io.Reader
}

func NewMultiReader() *MultiReader {
	return &MultiReader{
		bodyGetFunc: func() io.Reader {
			return NoBody
		},
		bodyLength: int64(0),
	}
}

func (slf *MultiReader) Close() error {
	if nil != slf.bodyRaw {
		return slf.bodyRaw.Close()
	} else {
		return nil
	}
}

func (slf *MultiReader) GetBody() io.Reader {
	return slf.bodyGetFunc()
}

func (slf *MultiReader) SetBody(body io.Reader) error {
	if nil == body {
		return errors.New("nil body")
	}
	if rc, ok := body.(io.ReadCloser); !ok && body != nil {
		slf.bodyRaw = ioutil.NopCloser(body)
	} else {
		slf.bodyRaw = rc
	}

	switch value := body.(type) {
	case *bytes.Buffer:
		slf.bodyLength = int64(value.Len())
		buf := value.Bytes()
		slf.bodyGetFunc = func() io.Reader {
			r := bytes.NewReader(buf)
			return ioutil.NopCloser(r)
		}
	case *bytes.Reader:
		slf.bodyLength = int64(value.Len())
		snapshot := *value
		slf.bodyGetFunc = func() io.Reader {
			r := snapshot
			return ioutil.NopCloser(&r)
		}
	case *strings.Reader:
		slf.bodyLength = int64(value.Len())
		snapshot := *value
		slf.bodyGetFunc = func() io.Reader {
			r := snapshot
			return ioutil.NopCloser(&r)
		}
	default:
		if nil != body {
			return errors.New("unknown body reader implements: " + reflect.TypeOf(body).String())
		}
	}
	return nil
}

////

var NoBody = noBody{}

type noBody struct{}

func (noBody) Read([]byte) (int, error)         { return 0, io.EOF }
func (noBody) Close() error                     { return nil }
func (noBody) WriteTo(io.Writer) (int64, error) { return 0, nil }
