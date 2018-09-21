package splitter

import (
	"bufio"
	"github.com/pkg/errors"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 按行分隔的Splitter
type LineScannerSplitter struct{}

func (slf *LineScannerSplitter) Split(source interface{}, events *SplitterEvents) error {
	if events == nil {
		panic("Events is nil")
	}

	if events.OnReceived == nil {
		panic("Events.OnReceived func is nil")
	}

	reader, err := ConvertToBufferedReader(source)
	if nil != err {
		return err
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if events.OnReadStart != nil {
			events.OnReadStart()
		}
		events.OnReceived(scanner.Bytes())
	}

	if err := scanner.Err(); nil != err {
		return errors.WithStack(err)
	} else {
		return nil
	}
}
