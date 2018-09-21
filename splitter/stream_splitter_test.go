package splitter

import (
	"fmt"
	"strings"
	"testing"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func TestReaderDelimSplitter(t *testing.T) {
	reader := strings.NewReader(`
{"A":1}
{"B":2}
{"B":2}
{"B":2}
{"B":2}
`)
	splitter := new(StreamDelimitedSplitter)
	err := splitter.Split(reader, &SplitterEvents{
		OnReceived: func(bytes []byte) {
			fmt.Println("Text:" + string(bytes))
		},
	})
	if nil != err {
		t.Error(err.Error())
	}
}
