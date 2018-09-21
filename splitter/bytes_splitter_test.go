package splitter

import (
	"fmt"
	"os"
	"testing"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func TestABytesSplitter_String(t *testing.T) {
	source := "ABC"
	sp := new(BytesSplitter)
	err := sp.Split(source, &SplitterEvents{
		OnReceived: func(bytes []byte) {
			fmt.Println("Txt:", string(bytes))
		},
	})
	if err != nil {
		t.Error(err)
	}
}

func TestABytesSplitter_File(t *testing.T) {
	source, fe := os.Open("../test-data.txt")
	if nil != fe {
		t.Error(fe)
	}
	sp := new(BytesSplitter)
	err := sp.Split(source, &SplitterEvents{
		OnReceived: func(bytes []byte) {
			fmt.Println("Txt:", string(bytes))
		},
	})
	if err != nil {
		t.Error(err)
	}
}
