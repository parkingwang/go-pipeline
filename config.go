package gopl

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//
import (
	"github.com/parkingwang/go-conf"
)

////

const componentTypeFieldName = "component"

// 插件配置选项
type ComponentConfig struct {
	ComponentType string   `toml:"component"` // 插件类型名称: componentTypeFieldName
	Disabled      bool     `toml:"disabled"`  // 是否禁用此插件，默认为false，即不禁用
	Topic         string   `toml:"topic"`     // Topic名称。Input/Filter/Output插件使用此字段来匹配消息
	DecoderName   string   `toml:"decoder"`   // Decoder名称，Output插件使用此字段
	InitArgs      conf.Map `toml:"InitArgs"`  // 插件初始化参数
}

// 调试配置选项
type DebugConfig struct {
	Verbose         bool   `toml:"verbose"`            // 详细输出日志消息
	VeryVerbose     bool   `toml:"very_verbose"`       // 非常详细地输出日志
	PanicCoreError  bool   `toml:"panic_thread_error"` // 线程池内部错误，使用Panic输出。调试中使用
	RoutingTrace    bool   `toml:"routing_trace"`      // 输出消息跟踪情况
	BlockDetectTime string `toml:"block_detect_time"`  // 消息处理阻塞检测时间
}

// 获取全局Globals配置。
// 它对应着配置文件的 [Globals] 配置项。
func Globals() conf.Map {
	return SharedRouter().globalsConfig
}

// 获取全局 Debug 配置。
// 它对应着配置文件的 [Debug] 配置项。
func Debugs() DebugConfig {
	return SharedRouter().debugConfig
}

// FindConfigOnRoot 返回从根配置中查找指定命名的配置对象。
// 如果配置不存在，返回nil, false，否则为 config, true
func FindConfigOnRoot(name string) (conf.Map, bool) {
	cnf := SharedRouter().rootConfig.GetMapOrDefault(name, nil)
	return cnf, nil != cnf
}
