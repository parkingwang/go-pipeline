package http

import (
	"github.com/parkingwang/go-conf"
	"github.com/parkingwang/go-sign"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-realip"
	"github.com/yoojia/go-ttlmap"
	"net/http"
	"time"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

// 初始化授权数据
func LoadAuthorizedApps(config conf.Map) *AuthorizeApps {

	authApps := config.GetMapArrayOrDefault("Apps", make([]conf.Map, 0))
	if len(authApps) == 0 {
		log.Error().Msg("未找到授权列表：Apps")
	}

	whiteList := config.GetMapArrayOrDefault("WhiteList", make([]conf.Map, 0))
	if len(authApps) == 0 {
		log.Error().Msg("未找到白名单列表：WhiteList")
	}

	allowPrivateNetwork := config.GetBoolOrDefault("allow_private_net", true)

	return &AuthorizeApps{
		AllowPrivateNetwork: allowPrivateNetwork,
		Authorized:          parseItems(authApps, "app_key", "app_secret"),
		WhiteHosts:          parseItems(whiteList, "host", "remark"),
	}
}

func parseItems(items []conf.Map, kName, vName string) map[string]string {
	output := make(map[string]string)
	for _, item := range items {
		output[item.MustString(kName)] = item.MustString(vName)
	}
	return output
}

func NewAuthMiddleware(keepAuthTTL time.Duration, authApp *AuthorizeApps) func(http.ResponseWriter, *http.Request) bool {
	hostCache := ttlmap.New()

	for k := range authApp.Authorized {
		log.Debug().Msgf("Auth app: " + k)
	}

	for k := range authApp.WhiteHosts {
		log.Debug().Msgf("White host: " + k)
	}

	return func(writer http.ResponseWriter, r *http.Request) bool {
		// 内网
		clientHost := realip.FromRequest(r)
		if ok, _ := realip.IsPrivateAddress(clientHost); authApp.AllowPrivateNetwork && ok {
			return true
		}

		// 白名单
		if _, ok := authApp.WhiteHosts[clientHost]; ok {
			return true
		}

		// 检查校验时效
		if hostCache.Exists(clientHost) {
			return true
		}

		verifier := sign.NewGoVerifier()
		if err := r.ParseForm(); nil != err {
			log.Error().Msgf("客户端授权验证无法解析, ClientIp: %s", clientHost)
		} else {
			verifier.ParseValues(r.Form)
		}

		// Check Params
		if err := verifier.MustHasOtherKeys(); nil != err {
			log.Error().Msgf("客户端授权验证失败, ClientIp: %s, 参数不完整: %s", clientHost, err.Error())
			return false
		}

		appId := verifier.GetAppId()
		if "" == appId {
			log.Error().Msgf("客户端授权验证失败, ClientIp: %s, 缺少授权AppId", clientHost)
			return false
		}

		// 签名时间戳，5分钟超时
		verifier.SetTimeout(time.Minute * 5)
		// 检查时间戳
		if err := verifier.CheckTimeStamp(); nil != err {
			log.Error().Msgf("客户端授权验证失败, ClientIp: %s, 无效的时间戳: %s", clientHost, err.Error())
			return false
		}

		// 签名时间戳，5分钟超时
		verifier.SetTimeout(time.Minute * 5)
		// 检查时间戳
		if err := verifier.CheckTimeStamp(); nil != err {
			log.Error().Msgf("客户端授权验证失败, ClientIp: %s, 无效的时间戳: %s", clientHost, err.Error())
			return false
		}

		if appSecret, has := authApp.Authorized[appId]; !has {
			log.Error().Msgf("客户端授权验证失败, ClientIp: %s, 无效的AppId: %s", clientHost, appId)
			return false
		} else {
			reSigner := sign.NewGoSignerHmac()
			reSigner.SetAppSecret(appSecret)
			reSigner.SetBody(verifier.GetBodyWithoutSign())

			serviceSign := reSigner.GetSignature()
			clientSign := verifier.GetSign()

			if serviceSign != clientSign {
				log.Error().Msgf("客户端授权验证失败, 签名不一致。ClientIp: %s, 服务端签名: %s, 客户端: %s", clientHost, serviceSign, clientSign)
				return false
			} else {
				log.Debug().Msgf("客户端授权验证成功, ClientIp: %s, RequestUri: [%s]%s", clientHost, r.Method, r.RequestURI)
				// 校验成功后，设置无须校验的时效
				hostCache.Add(clientHost, 0, keepAuthTTL)
				return true
			}
		}
	}
}

////

type AuthorizeApps struct {
	AllowPrivateNetwork bool              // 允许局域网访问
	Authorized          map[string]string // 授权列表
	WhiteHosts          map[string]string // 白名单列表
}
