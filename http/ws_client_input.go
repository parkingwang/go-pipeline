package http

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

import (
	"github.com/gorilla/websocket"
	"github.com/parkingwang/go-conf"
	"github.com/parkingwang/go-sign"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-pipeline"
	"github.com/yoojia/go-pipeline/abc"
	"net/http"
	"time"
)

type GoPLWebSocketClientInput struct {
	gopl.AbcSlot
	abc.AbcShutdown

	serverPath              string
	clientReconnectInterval time.Duration
	clientReadTimeout       time.Duration
	clientOrigin            string
	authEnabled             bool
	authAppKey              string
	authAppSecret           string
}

func (slf *GoPLWebSocketClientInput) Init(args conf.Map) {
	slf.AbcSlot.Init(args)
	slf.AbcShutdown.Init()

	slf.serverPath = args.MustString("server_path")
	slf.clientReadTimeout = args.GetDurationOrDefault("client_read_timeout", time.Second*5)
	slf.clientReconnectInterval = args.GetDurationOrDefault("client_reconnect_interval", time.Second*5)
	slf.clientOrigin = args.GetStringOrDefault("client_origin", "gopl://websocket-client")
	slf.authEnabled = args.MustBool("auth_enabled")
	slf.authAppKey = args.MustString("auth_app_key")
	slf.authAppSecret = args.MustString("auth_app_secret")

	if slf.authEnabled {
		if "" == slf.authAppKey || "" == slf.authAppSecret {
			slf.TagLog(log.Panic).Msg("Auth enabled must set <auth_app_id> and <auth_app_secret>")
		}
	}
}

func (slf *GoPLWebSocketClientInput) Input(deliverer gopl.Deliverer, decoder gopl.Decoder) {
	defer slf.SetTerminated()

	header := http.Header{}
	header.Add("Origin", slf.clientOrigin)
	header.Add("X-Client", "GoPipeline/WebSocketClientInput")

	initUrl := slf.makeWSUrl()
	cli, _, err := websocket.DefaultDialer.Dial(initUrl, header)
	if nil != err {
		slf.TagLog(log.Error).Err(err).Msgf("Dial to server: %s FAILED", initUrl)
	} else {
		slf.TagLog(log.Info).Msgf("Dial to server: %s SUCCESS", initUrl)
	}

	defer func() {
		if nil != cli {
			cli.Close()
		}
	}()

	reConnTicker := time.NewTicker(slf.clientReconnectInterval)
	defer reConnTicker.Stop()

	for {
		select {
		case <-slf.ShutdownChan():
			return

		case <-reConnTicker.C:
			if nil == cli {
				reconnectUrl := slf.makeWSUrl()
				slf.TagLog(log.Info).Msgf("Redial to server: %s", reconnectUrl)
				cli, _, err = websocket.DefaultDialer.Dial(reconnectUrl, header)
				if nil != err {
					slf.TagLog(log.Error).Err(err).Msg("Redial to server: FAILED")
				} else {
					slf.TagLog(log.Info).Msgf("Redial to server: %s SUCCESS", reconnectUrl)
				}
			}

		default:
			if nil != cli {
				cli.SetReadDeadline(time.Now().Add(slf.clientReadTimeout))
				if _, bytes, err := cli.ReadMessage(); nil != err {
					slf.TagLog(log.Error).Err(err).Msgf("Read from server: %s FAILED", slf.serverPath)
					cli.Close()
					cli = nil
				} else {
					if msg, err := decoder.Decode(bytes); nil != err {
						slf.TagLog(log.Error).Err(err).Str("bytes", string(bytes)).Msgf("Decode ws bytes FAILED")
					} else {
						deliverer.Deliver(msg)
					}
				}
			}
		}
	}
}

func (slf *GoPLWebSocketClientInput) makeWSUrl() string {
	url := slf.serverPath
	if slf.authEnabled {
		signer := sign.NewGoSignerHmac()
		signer.SetAppId(slf.authAppKey)
		signer.SetAppSecret(slf.authAppSecret)
		signer.RandNonceStr()
		signer.SetTimeStamp(time.Now().Unix())
		url = url + "?" + signer.GetSignedQuery()
	}
	return url
}
