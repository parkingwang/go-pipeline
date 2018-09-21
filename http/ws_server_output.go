package http

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-pipeline"
	"github.com/yoojia/go-pipeline/abc"
	"github.com/yoojia/go-pipeline/util"
	"math"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
// WebSocket服务端输出插件。根据配置，可开启Origin校验功能来限制连接目标。
// 服务端向所有空闲的客户端广播输出消息。
//   - 服务端为每个客户端配置一个指定最大容量的消息缓存池；
//	 - 客户端从缓存池中读取消息数据；

type GoPLWebSocketServerOutput struct {
	gopl.AbcSlot
	abc.AbcShutdown

	upgrader     *websocket.Upgrader // WebSocket Upgrader
	writeTimeout time.Duration       // WebSocket向客户端写入数据的超时时间
	pathUri      string              // WebSocket连接地址
	cliCacheSize int                 // 客户端消息缓存池容量
	cliMaxCount  int32               // WS服务端最大连接的客户端数量
	cliNowCount  int32               // WS服务端已经连接的客户端数量
	sessions     *sync.Map           // 所有连接的客户
}

func (slf *GoPLWebSocketServerOutput) Init(args conf.Map) {
	slf.AbcSlot.Init(args)
	slf.AbcShutdown.Init()

	slf.writeTimeout = gopl.DurationOrDefault(args.MustString("conn_write_timeout"), time.Second*3)
	slf.pathUri = args.GetStringOrDefault("path_uri", "/ws")
	size := args.GetInt64OrDefault("session_cache_size", 1)
	slf.cliCacheSize = util.MaxInt(int(size), 1)

	// Check Origin default
	authCheckOrigin := args.GetBoolOrDefault("auth_check_origin", true)
	authOrigins, _ := args.MustStringArray("auth_origins")
	slf.upgrader = &websocket.Upgrader{
		HandshakeTimeout: gopl.DurationOrDefault(args.MustString("conn_handshake_timeout"), time.Second*5),
		CheckOrigin: func(r *http.Request) bool {
			if !authCheckOrigin {
				return true
			}
			origin := r.Header.Get("Origin")
			for _, auth := range authOrigins {
				if auth == origin {
					return true
				}
			}
			slf.TagLog(log.Info).Msgf("Check origin: NOT-MATCH, was: %s", origin)
			return false
		},
	}

	slf.sessions = new(sync.Map)

	max := args.GetInt64OrDefault("server_max_clients", 64)
	slf.cliMaxCount = util.MaxInt32(int32(max), 1)

	slf.TagLog(log.Info).Msg("WebSocket server initialize")
	go slf.onServe()
}

func (slf *GoPLWebSocketServerOutput) Output(pack *gopl.DataFrame) {
	if slf.cliNowCount <= 0 {
		return
	}
	if bytes, err := pack.ReadBytes(); nil != err {
		slf.TagLog(log.Error).Err(err).Str("raw", fmt.Sprintf("%s", pack)).Msg("Read bytes FAILED")
	} else {
		slf.forEachClients(func(addr string, cli *WsSession) {
			select {
			case cli.Send() <- bytes:
			default:
				slf.TagLog(log.Error).Msgf("Client BLOCKED: %s", addr)
			}
		})
	}
}

func (slf *GoPLWebSocketServerOutput) onServe() {
	defer slf.SetTerminated()

	RegisterHandler("GET", slf.pathUri, func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		if conn, err := slf.upgrader.Upgrade(w, r, nil); err != nil {
			slf.TagLog(log.Error).Err(err).Msgf("Upgrade WS protocol FAILED: %s", r.RemoteAddr)
		} else {
			slf.onClientOpened(conn)
		}
	})
	slf.TagLog(log.Info).Msgf("Register web-socket serve on: %s", slf.pathUri)

	<-slf.ShutdownChan()
	atomic.StoreInt32(&slf.cliNowCount, math.MinInt32)
	slf.TagLog(log.Info).Msgf("Shutdown, close sessions...")
	slf.forEachClients(func(addr string, cli *WsSession) {
		cli.CloseSendBuffer()
	})
	slf.TagLog(log.Info).Msgf("Shutdown, close sessions: OK")
}

func (slf *GoPLWebSocketServerOutput) onClientOpened(conn *websocket.Conn) {
	addr := conn.RemoteAddr().String()
	// 客户端数量为负，服务器正在关闭
	if slf.cliNowCount < 0 {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"DEAD_SERVER"}`))
		slf.TagLog(log.Debug).Msgf("Client[WS:%s] is REJECTED: DEAD_SERVER", addr)
		conn.Close()
		return
	}

	cli := NewWsSession(addr, conn, slf.writeTimeout, slf.cliCacheSize)

	go func() {
		// 客户端在Run函数中持续运行。
		cli.Run()
		// 如果函数返回，则说明客户端已中断/停止
		slf.TagLog(log.Debug).Msgf("Client[WS:%s] is CLOSED", cli.addr)
		slf.sessions.Delete(cli.addr)
		atomic.AddInt32(&slf.cliNowCount, -1)
	}()

	slf.sessions.Store(addr, cli)
	atomic.AddInt32(&slf.cliNowCount, 1)

}

func (slf *GoPLWebSocketServerOutput) forEachClients(fn func(addr string, cli *WsSession)) {
	slf.sessions.Range(func(k, v interface{}) bool {
		fn(k.(string), v.(*WsSession))
		return true
	})
}

////

type WsSession struct {
	opts       *WsOption
	addr       string
	wsConn     *websocket.Conn
	sendBuffer chan []byte
	sendClosed bool
}

func NewWsSession(addr string, conn *websocket.Conn, timeout time.Duration, cacheSize int) *WsSession {
	return &WsSession{
		addr: addr,
		opts: &WsOption{
			DataType:     websocket.TextMessage,
			WriteTimeout: timeout,
		},
		wsConn:     conn,
		sendBuffer: make(chan []byte, cacheSize),
		sendClosed: false,
	}
}

func (slf *WsSession) CloseSendBuffer() {
	if !slf.sendClosed {
		close(slf.sendBuffer)
		slf.sendClosed = true
	}
}

func (slf *WsSession) Send() chan<- []byte {
	return slf.sendBuffer
}

func (slf *WsSession) Run() {
	// Send loop
	for bytes := range slf.sendBuffer {
		slf.wsConn.SetWriteDeadline(time.Now().Add(slf.opts.WriteTimeout))
		if err := slf.wsConn.WriteMessage(slf.opts.DataType, bytes); err != nil {
			log.Error().Err(err).Msgf("Client[WS:%s] write FAILED", slf.addr)
			slf.CloseSendBuffer() // Close函数会中断Send循环
		}
	}
	slf.wsConn.WriteMessage(websocket.CloseMessage, []byte{})
	slf.wsConn.Close()
}

type WsOption struct {
	DataType     int
	WriteTimeout time.Duration
}
