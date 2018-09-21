package gopl

import (
	"bytes"
	"container/list"
	"fmt"
	"github.com/parkingwang/go-conf"
	"github.com/pelletier/go-toml"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-goes"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"
)

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

type GoPipeline struct {
	rootConfig           conf.Map
	globalsConfig        conf.Map
	debugConfig          DebugConfig
	debugDetectBlockTime time.Duration

	startupHook  *list.List
	shutdownHook *list.List

	inputRunners  *list.List
	filterRunners *list.List
	outputRunners *list.List

	plugins *list.List

	threads *goes.GoesPool
	signals chan os.Signal

	decoders map[string]Decoder
	matchers map[string]Matcher

	factoryInputs  map[string]InputFactory
	factoryOutputs map[string]OutputFactory
	factoryFilters map[string]FilterFactory
}

var gSharedRouter = newRouter(runtime.NumCPU() * 8)

func SharedRouter() *GoPipeline {
	return gSharedRouter
}

// 加载预设组件
func (slf *GoPipeline) Prepare(prepares ...func(router *GoPipeline)) {
	slf.AutoRegister(new(JSONDecoder))

	for _, prepare := range prepares {
		prepare(slf)
	}
}

// 初始化
func (slf *GoPipeline) Setup(path string) {
	// 初始化全局配置
	slf.setupConfig(path)

	// Log registered items
	for dn := range slf.decoders {
		withTag(log.Info).Msgf("Registered Decoder: <%s>", dn)
	}
	for mn := range slf.matchers {
		withTag(log.Info).Msgf("Registered Matcher: <%s>", mn)
	}
	for in := range slf.factoryInputs {
		withTag(log.Info).Msgf("Registered Input: <%s>", in)
	}
	for fn := range slf.factoryFilters {
		withTag(log.Info).Msgf("Registered Filter: <%s>", fn)
	}
	for on := range slf.factoryOutputs {
		withTag(log.Info).Msgf("Registered Output: <%s>", on)
	}

	ifComponentEnabled := func(configKey string) (*ComponentConfig, bool) {
		config := ComponentConfig{}
		if m, ok := slf.rootConfig[configKey]; !ok {
			return nil, false
		} else {
			if err := conf.Map2Struct(m, &config); nil != err {
				withTag(log.Error).Err(err).Msg("Decode map to ComponentConfig FAILED")
				return nil, false
			} else {
				// 默认的插件类型为当前的配置名。
				// 即实现插件类型名即可作为配置名；而同时配置同一类型插件时，需要使用不同的配置名。
				if "" == config.ComponentType {
					config.ComponentType = configKey
				}
			}
		}

		// 在配置文件中未定义，或者设置为Disabled状态的组件，不注册
		if config.Disabled {
			withTag(log.Warn).Msgf("Component[%s] define in .toml, but is set to <DISABLED>", configKey)
			return nil, false
		}

		return &config, true
	}

	// 组件列表
	for componentKey, val := range slf.rootConfig {
		// 判断是否为组件配置字段
		config, is := val.(map[string]interface{})
		if !is {
			continue
		}
		cType, cTypeName, ok := ifCoreComponentConfig(componentKey, config)
		if !ok {
			continue
		}

		// 根据解析出来的参数，注册组件到内核
		switch cType {
		case componentInput:
			if config, en := ifComponentEnabled(componentKey); en {
				factory, _ := slf.factoryInputs[cTypeName]
				input := factory()
				decoder := findNonNilDecoder(input, config, componentKey)
				slf.inputRunners.PushBack(newInputRunner(input, decoder, config, componentKey))
				withTag(log.Info).Msgf("Working Input: <%s>", componentKey)
			}

		case componentFilter:
			if cnf, en := ifComponentEnabled(componentKey); en {
				factory, _ := slf.factoryFilters[cTypeName]
				filter := factory()
				matcher := findNonNilMatcher(filter, cnf)
				fr := newFilterRunner(filter, matcher, cnf, componentKey)
				slf.filterRunners.PushBack(fr)
				withTag(log.Info).Msgf("Working Filter: <%s>", componentKey)
			}

		case componentOutput:
			if cnf, en := ifComponentEnabled(componentKey); en {
				newOutputFactory, _ := slf.factoryOutputs[cTypeName]
				output := newOutputFactory()
				matcher := findNonNilMatcher(output, cnf)
				or := newOutputRunner(output, matcher, cnf, componentKey)
				slf.outputRunners.PushBack(or)
				withTag(log.Info).Msgf("Working Output: <%s>", componentKey)
			}
		}
	}

}

func (slf *GoPipeline) Init() {
	// Init components
	withTag(log.Info).Msg("Initialing components")

	// 插件
	for ele := slf.plugins.Front(); ele != nil; ele = ele.Next() {
		detectTimeout(ele.Value.(Plugin).Init, "Plugin.Init")
	}
	// Outputs
	for ele := slf.outputRunners.Front(); ele != nil; ele = ele.Next() {
		detectTimeout(ele.Value.(*outputRunner).init, "OutputRunner.Init")
	}
	// Filters
	for ele := slf.filterRunners.Front(); ele != nil; ele = ele.Next() {
		detectTimeout(func() {
			ele.Value.(*filterRunner).init(slf)
		}, "FilterRunner.Init")
	}
	// Inputs
	for ele := slf.inputRunners.Front(); ele != nil; ele = ele.Next() {
		detectTimeout(ele.Value.(*inputRunner).init, "InputRunner.Init")
	}
}

func (slf *GoPipeline) setupConfig(dirpath string) {
	if "" == dirpath {
		dirpath = "conf.d"
	}

	if fi, err := os.Stat(dirpath); nil != err || !fi.IsDir() {
		withTag(log.Panic).Err(err).Msgf("Config path muse be a dir: %s", dirpath)
		return
	}

	mergedTxt := new(bytes.Buffer)
	withTag(log.Info).Msgf("Load config dir: %s", dirpath)
	if files, err := ioutil.ReadDir(dirpath); nil != err {
		withTag(log.Panic).Err(err).Msgf("Failed to list file in dir: %s", dirpath)
	} else {
		if 0 == len(files) {
			withTag(log.Panic).Err(err).Msgf("Config file NOT FOUND in dir: %s", dirpath)
		}
		for _, f := range files {
			name := f.Name()
			if !strings.HasSuffix(name, ".toml") {
				continue
			}
			path := fmt.Sprintf("%s%s%s", dirpath, "/", f.Name())
			withTag(log.Info).Msgf("Load config file: %s", path)
			if bs, err := ioutil.ReadFile(path); nil != err {
				withTag(log.Panic).Err(err).Msgf("Failed to load file: %s", path)
			} else {
				mergedTxt.Write(bs)
			}
		}
	}

	if tree, err := toml.LoadBytes(mergedTxt.Bytes()); nil != err {
		withTag(log.Panic).Err(err).Msg("Failed to decode toml config file")
	} else {
		slf.rootConfig = tree.ToMap()
	}

	if len(slf.rootConfig) == 0 {
		withTag(log.Panic).Msg("Root config is EMPTY")
	}

	// Globals args
	slf.globalsConfig = slf.rootConfig.MustMap("Globals")
	// Debugs config
	debug := slf.rootConfig.MustMap("Debug")
	if len(debug) > 0 {
		if err := conf.Map2Struct(debug, &slf.debugConfig); nil != err {
			withTag(log.Panic).Err(err).Msg("Failed to decode map to [Debug] config")
		}
	}
}

func (slf *GoPipeline) startup() {
	// 启动
	resetFioCounter()
	// Core Threads
	slf.threads.Start()

	// 组件最先启动
	for ele := slf.plugins.Front(); ele != nil; ele = ele.Next() {
		detectTimeout(ele.Value.(Plugin).Startup, "Plugin.Startup")
	}
	// Hooks
	for ele := slf.startupHook.Front(); ele != nil; ele = ele.Next() {
		detectTimeout(ele.Value.(HookFunc), "StartupHooks.Call")
	}
	// Start Inputs
	for ele := slf.inputRunners.Front(); ele != nil; ele = ele.Next() {
		detectTimeout(func() {
			go ele.Value.(*inputRunner).start(slf)
		}, "InputRunner.Start")
	}
}

func (slf *GoPipeline) shutdown() {
	// 停止
	closeSlot := func(slot VirtualSlot) {
		if shutdown, ok := slot.(NeedShutdown); ok {
			withTag(log.Info).Msgf("Shutdown slot: %s", slot.GetName())
			shutdown.Shutdown()
			withTag(log.Info).Msgf("Shutdown slot: %s [COMPLETED]", slot.GetName())
		}
	}
	withTag(log.Info).Msg("Shutdown components...")
	// Inputs
	for ele := slf.inputRunners.Back(); ele != nil; ele = ele.Prev() {
		closeSlot(ele.Value.(*inputRunner).input)
	}
	// Filters
	for ele := slf.filterRunners.Back(); ele != nil; ele = ele.Prev() {
		closeSlot(ele.Value.(*filterRunner).filter)
	}
	// Outputs
	for ele := slf.outputRunners.Back(); ele != nil; ele = ele.Prev() {
		closeSlot(ele.Value.(*outputRunner).output)
	}
	withTag(log.Info).Msg("Shutdown components: [COMPLETED]")

	// Hooks
	for ele := slf.shutdownHook.Back(); ele != nil; ele = ele.Prev() {
		ele.Value.(HookFunc)()
	}
	// Plugin
	for ele := slf.plugins.Back(); ele != nil; ele = ele.Prev() {
		ele.Value.(Plugin).Shutdown()
	}
	// Core Threads
	slf.threads.Shutdown()
}

// 启动消息路由
func (slf *GoPipeline) StartRoute() {
	// 接收系统中断信号
	signal.Notify(slf.signals, os.Interrupt, os.Kill)

	slf.startup()
	withTag(log.Info).Msg("STARTED")
	defer withTag(log.Info).Msg("STOPPED")

	// 等待系统中断信号
	switch <-slf.signals {
	case os.Kill:
		withTag(log.Info).Msg("Received system [KILL] signal")
		t := time.AfterFunc(time.Second*5, func() {
			withTag(log.Info).Msg("Shutdown failed(3s timeout), FORCE KILL !!")
			os.Exit(-1)
		})
		slf.shutdown()
		t.Stop()

	case os.Interrupt:
		withTag(log.Info).Msg("Received system [INTERRUPT] signal")
		slf.shutdown()
	}

}

// 停止
func (slf *GoPipeline) StopRouter() {
	slf.signals <- os.Interrupt
}

// 接收到消息投递
func (slf *GoPipeline) Deliver(pack *DataFrame) {
	// 监控每个消息的处理阻塞情况，如果超时未完成处理，则输出警告信息。
	// TODO 有没有更好的方法来处理？
	t := time.AfterFunc(slf.debugDetectBlockTime, func() {
		withTag(log.Warn).Msgf("Deliver BLOCKED: timeout > %s, sender: %s", slf.debugDetectBlockTime, pack.Sender())
	})

	// 使用协程池来派发消息
	slf.threads.Post(func() {
		defer t.Stop()
		slf.deliver0(pack)
	})
}

func (slf *GoPipeline) deliver0(pack *DataFrame) {

	defer func() {
		if r := recover(); nil != r {
			if err, ok := r.(error); ok {
				withTag(log.Error).Err(err).Msg("Error in core-goroutine")
			}

			stackBuf := make([]byte, 1024*4)
			stackBuf = stackBuf[:runtime.Stack(stackBuf, false)]
			withTag(log.Error).Str("stack", string(stackBuf)).Msg("Goroutine error")

			if slf.debugConfig.PanicCoreError {
				panic(r)
			}
		}
	}()

	// 缓存Filter处理结果，容量是Filter的数量+1
	filteredOut := make([]*DataFrame, 1, slf.filterRunners.Len()+1)
	// 第一个缓存是当前等待处理的消息，其它是Filter的输出结果
	filteredOut[0] = pack

	// filter
	for ele := slf.filterRunners.Front(); ele != nil; ele = ele.Next() {
		fr := ele.Value.(*filterRunner)
		if !fr.checkAccept(pack) {
			if slf.debugConfig.VeryVerbose {
				withTag(log.Debug).Msgf("REJECTED [xx] Filter: <%s> , sender: %s", fr.filter.GetName(), pack.Sender())
			}
			continue
		}

		if slf.debugConfig.RoutingTrace {
			withTag(log.Debug).Msgf("ACCEPTED [√√] Filter: <%s> , sender: %s", fr.filter.GetName(), pack.Sender())
		}

		s1 := time.Now()
		// Filter返回输出消息。如果不是原样返回，则交给Output来处理。
		if ret := fr.runFilter(s1, pack); nil != ret && pack != ret {
			filteredOut = append(filteredOut, ret)
		}
		// 统计采样Filter处理消息的耗时
		takes := time.Now().Sub(s1)
		if takes >= slf.debugDetectBlockTime {
			withTag(log.Warn).Msgf("Filter: <%s> BLOCKED, takes: %s", fr.filter.GetName(), takes)
		}
		// Counting & Samples
		go func() {
			increaseFilter()
			sampleFilter(takes.Nanoseconds())
		}()
	}

	// finally release all messages
	defer func() {
		for _, r := range filteredOut {
			if nil == r {
				break
			}
			releaseDataFrame(r)
		}
	}()

	// Output
	for _, ret := range filteredOut {
		if nil == ret {
			break
		}
		for ele := slf.outputRunners.Front(); ele != nil; ele = ele.Next() {
			or := ele.Value.(*outputRunner)

			if !or.checkAccept(ret) {
				if slf.debugConfig.VeryVerbose {
					withTag(log.Debug).Msgf("REJECTED [xx] Output: <%s>, sender: %s", or.output.GetName(), ret.Sender())
				}
				continue
			}

			if slf.debugConfig.RoutingTrace {
				withTag(log.Debug).Msgf("ACCEPTED [√√] Output: <%s> , sender: %s", or.output.GetName(), ret.Sender())
			}

			s2 := time.Now()
			or.runOutput(ret)
			// 统计采样Output处理消息的耗时
			takes := time.Now().Sub(s2)
			if takes >= slf.debugDetectBlockTime {
				withTag(log.Warn).Msgf("Output: <%s> BLOCKED, takes: %s", or.output.GetName(), takes)
			}
			// Counting & Samples
			go func() {
				increaseOutbounds()
				sampleOutbounds(takes.Nanoseconds())
			}()
		}
	}
}

// 创建Router，指定处理消息的协程最大数量
func newRouter(maxGoNum int) *GoPipeline {
	return &GoPipeline{
		rootConfig:           conf.Map{},
		globalsConfig:        conf.Map{},
		debugConfig:          DebugConfig{},
		debugDetectBlockTime: time.Millisecond * 500,

		startupHook:   list.New(),
		shutdownHook:  list.New(),
		inputRunners:  list.New(),
		filterRunners: list.New(),
		outputRunners: list.New(),
		plugins:       list.New(),

		decoders:       make(map[string]Decoder),
		matchers:       make(map[string]Matcher),
		factoryInputs:  make(map[string]InputFactory),
		factoryOutputs: make(map[string]OutputFactory),
		factoryFilters: make(map[string]FilterFactory),

		threads: goes.NewGoesPool(maxGoNum, maxGoNum),
		signals: make(chan os.Signal, 1),
	}
}

func detectTimeout(actionFunc func(), action string) {
	t := time.AfterFunc(time.Second, func() {
		withTag(log.Warn).Msgf("%s takes TOO LOOONG: >1s", action)
	})
	defer t.Stop()

	actionFunc()
}

func withTag(f func() *zerolog.Event) *zerolog.Event {
	return f().Str("tag", "GoPipeline")
}
