package common

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//
import (
	"bytes"
	"fmt"
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-jsonx"
	"github.com/yoojia/go-pipeline"
)

type GoPLConsoleOutput struct {
	gopl.AbcSlot
}

func (slf *GoPLConsoleOutput) Init(args conf.Map) {
	slf.AbcSlot.Init(args)
}

func (slf *GoPLConsoleOutput) Output(pack *gopl.DataFrame) {
	out := bytes.NewBuffer(make([]byte, 0))
	err := jsonx.CompressJSON(pack.GetBody(), out)
	if nil != err {
		if err == jsonx.ErrNotJSONData {
			slf.TagLog(log.Debug).Str("txt", out.String()).Msg("Output:String")
		} else {
			slf.TagLog(log.Error).Err(err).Str("raw", fmt.Sprintf("%s", pack)).Msg(err.Error())
		}
	} else {
		slf.TagLog(log.Debug).RawJSON("json", out.Bytes()).Msg("Output:JSON")
	}
}
