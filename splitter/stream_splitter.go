package splitter

import (
	"github.com/pkg/errors"
	"io"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 字节流分隔符Splitter
type StreamDelimitedSplitter struct{}

func (slf *StreamDelimitedSplitter) Split(source interface{}, events *SplitterEvents) error {
	if events == nil {
		panic("Events is nil")
	}

	if events.OnReceived == nil {
		panic("Events.OnReceived func is nil")
	}

	delimiter := DataBytesDelimiter
	if events.Delimiter != byte(0) {
		delimiter = events.Delimiter
	}

	reader, err := ConvertToBufferedReader(source)
	if nil != err {
		return err
	}

	for {
		if events.OnReadStart != nil {
			events.OnReadStart()
		}
		if buf, err := reader.ReadBytes(delimiter); nil != err {
			// 数据流正常结束
			if io.EOF == err {
				return nil
			}

			// 没有错误检测函数，默认发生错误时中断内部循环
			if events.OnErrors == nil || events.OnErrors(err) {
				return errors.WithStack(err)
			}

		} else {
			if bytes, size := UnwrapDelimiter(delimiter, buf); size > 0 {
				events.OnReceived(bytes)
			}
		}
	}
}
