package http

import (
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-http"
	"github.com/yoojia/go-pipeline"
	"net/http"
	"time"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

var gHttpServer = httpd.NewHttpServer()
var gHttpRouter = httpd.NewHttpRouter()

// HttpRouter 返回用于注册Http处理函数路由的对象。
// 注意：获取路由对象时，Http服务不一定处于运行状态。
// 在非运行状态下注册的路由信息，需要等待Http服务运行后才生效。
func HttpRouter() *httpd.HttpRouter {
	return gHttpRouter
}

// RegisterHandler 用来注册Http处理函数，指定Pattern
func RegisterHandler(method, pattern string, handler httprouter.Handle) {
	HttpRouter().Handle(method, pattern, handler)
}

func ServerStartupHook() {
	httpConfig, hit := gopl.FindConfigOnRoot("GoPLHttpServer")
	if !hit {
		return
	}
	if httpConfig.GetBoolOrDefault("disabled", false) {
		return
	}

	address := httpConfig.GetStringOrDefault("address", ":18880")

	log.Info().Str("tag", "HttpServerHook").Msgf("Start Http Server, address: %s", address)
	go func() {

		if httpConfig.MustBool("auth_enabled") {
			log.Info().Str("tag", "HttpServerHook").Msg("Http Server: Auth ENABLED")
			authApps := LoadAuthorizedApps(httpConfig)
			authTTL := httpConfig.GetDurationOrDefault("auth_keep_ttl", time.Minute*10)
			HttpRouter().UseInterceptor(NewAuthMiddleware(authTTL, authApps))
		}

		err := gHttpServer.Start(address, HttpRouter().Route())
		if nil != err {
			if err != http.ErrServerClosed {
				log.Fatal().Err(err).Msg("Failed to start http server")
			}
			return
		}
	}()

	// 等待服务器启动
	<-time.After(time.Millisecond)
}

func ServerTerminateHook() {
	if gHttpServer.IsRunning() {
		if err := gHttpServer.Shutdown(); nil != err {
			log.Fatal().Err(err).Msg("Failed to shutdown http server")
		}
	}
}
