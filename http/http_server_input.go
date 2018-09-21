package http

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
// 启动Http服务端，监听POST端口接受客户端的输入
//

import (
	"fmt"
	"github.com/julienschmidt/httprouter"
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-jsonx"
	"github.com/yoojia/go-pipeline"
	"github.com/yoojia/go-pipeline/abc"
	"io/ioutil"
	"net/http"
	"strings"
)

type GoPLHttpServerInput struct {
	gopl.AbcSlot
	abc.AbcShutdown

	pathUri         string // 接收输入的消息的Http路径
	responseSuccess string // 响应给输入消息的Response消息
	responseFailed  string // 响应给输入消息的Response消息
}

func (slf *GoPLHttpServerInput) Init(args conf.Map) {
	slf.AbcSlot.Init(args)
	slf.AbcShutdown.Init()
	slf.pathUri = args.GetStringOrDefault("path_uri", "/")
	slf.responseSuccess = args.GetStringOrDefault("response_success", `{"message": "ok", "status": "success"}`)
	slf.responseFailed = args.GetStringOrDefault("response_failed", `{"message": "%s", "status": "fail"}`)
}

func (slf *GoPLHttpServerInput) Input(deliverer gopl.Deliverer, decoder gopl.Decoder) {
	defer slf.SetTerminated()

	// 处理每个Http请求
	RegisterHandler("POST", slf.pathUri, func(resp http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		defer req.Body.Close()

		resp.Header().Set("X-Server", "GoPipeline/HttpServerInput")
		resp.Header().Set("Content-Type", "application/json")

		sendResponseFailed := func(err string) {
			resp.WriteHeader(http.StatusBadRequest)
			resp.Write([]byte(fmt.Sprintf(slf.responseFailed, err)))
		}

		req.ParseForm()

		// PostForm数据
		bytes := make([]byte, 0)
		if len(req.PostForm) > 0 {
			json := jsonx.NewFatJSON()
			for key := range req.PostForm {
				txt := strings.TrimSpace(req.PostFormValue(key))
				if jsonx.HasJSONMark([]byte(txt)) {
					json.FieldNotEscapeValue(key, txt)
				} else {
					json.Field(key, txt)
				}
			}
			bytes = json.Bytes()
		} else { // Body数据
			if bs, err := ioutil.ReadAll(req.Body); nil != err {
				slf.TagLog(log.Error).Err(err).Msgf("Reading body FAILED")
				sendResponseFailed(err.Error())
				return
			} else {
				bytes = bs
			}
		}

		if 0 < len(bytes) {
			pack, err := decoder.Decode(bytes)
			if nil != err {
				slf.TagLog(log.Error).Err(err).Str("body", string(bytes)).Msg("Decode body FAILED")
				sendResponseFailed(err.Error())
				return
			}
			deliverer.Deliver(pack)
		}

		resp.WriteHeader(http.StatusOK)
		resp.Write([]byte(slf.responseSuccess))
	})

	// API校验

	slf.TagLog(log.Info).Msgf("Register http handler on: %s", slf.pathUri)
	// 保持持续运行，监听Shutdown信号
	<-slf.ShutdownChan()
}
