package gopl

import (
	"fmt"
	"github.com/parkingwang/go-conf"
	"testing"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

func TestMarshalJSON(t *testing.T) {
	txt := `
{
"A": 123,
"B": "yoojia"
}`
	out := conf.Map{}
	err := UnmarshalJSON([]byte(txt), &out)
	if nil != err {
		t.Error(err)
	}
	fmt.Println(out)
}
