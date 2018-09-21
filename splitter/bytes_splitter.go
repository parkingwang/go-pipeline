package splitter

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"os"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 读取中所有数据的Splitter
type BytesSplitter struct{}

func (slf *BytesSplitter) Split(source interface{}, events *SplitterEvents) error {
	if events == nil {
		panic("Events is nil")
	}
	if events.OnReceived == nil {
		panic("Events.OnReceived func is nil")
	}

	buffer := bytes.NewBuffer(make([]byte, 0))

	if events.OnReadStart != nil {
		events.OnReadStart()
	}

	switch source.(type) {
	case *os.File:
		f := source.(*os.File)
		if err := slf.read(buffer, bufio.NewReader(f)); nil != err {
			return errors.WithStack(err)
		}

	case string:
		buffer.WriteString(source.(string))

	case io.Reader:
		ior := source.(io.Reader)
		if err := slf.read(buffer, bufio.NewReader(ior)); nil != err {
			return errors.WithStack(err)
		}

	case *bufio.Reader:
		if err := slf.read(buffer, bufio.NewReader(source.(*bufio.Reader))); nil != err {
			return errors.WithStack(err)
		}

	default:
		return fmt.Errorf("unsupported source type, was: %t", source)
	}

	events.OnReceived(buffer.Bytes())

	return nil
}

func (slf *BytesSplitter) read(buffer *bytes.Buffer, reader *bufio.Reader) error {
	part := make([]byte, 1024)
	for {
		count, err := reader.Read(part)
		if nil != err && err != io.EOF {
			return errors.WithStack(err)
		}

		if count == 0 {
			return nil
		}

		buffer.Write(part[:count])
	}
}
