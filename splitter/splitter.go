package splitter

import (
	"bufio"
	"github.com/pkg/errors"
	"io"
	"os"
	"strings"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

type SplitterEvents struct {
	Delimiter   byte                 // Optional 数据流分割符号，默认为 \n 符号
	OnReceived  func([]byte)         // Required 处理接收到的数据。
	OnErrors    func(err error) bool // Optional 内部循环错误检测函数，返回 True 即中断内部循环。
	OnReadStart func()               // Optional 在每次读取数据循环开始时回调。
}

type Splitter interface {
	// 循环地分割数据源，使用Events的函数来处理。如果发生错误，返回出错对象。如果读完数据，返回nil。
	Split(source interface{}, events *SplitterEvents) error
}

func ConvertToBufferedReader(source interface{}) (*bufio.Reader, error) {
	var bufReader *bufio.Reader
	switch source.(type) {
	case *os.File:
		bufReader = bufio.NewReader(source.(*os.File))

	case string:
		bufReader = bufio.NewReader(strings.NewReader(source.(string)))

	case io.Reader:
		bufReader = bufio.NewReader(source.(io.Reader))

	case *bufio.Reader:
		bufReader = source.(*bufio.Reader)

	default:
		return nil, errors.Errorf("unsupported source type, was: %t", source)
	}

	return bufReader, nil
}
