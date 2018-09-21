package main

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//
import (
	"flag"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-pid"
	"github.com/yoojia/go-pipeline"
	"github.com/yoojia/go-pipeline/common"
	"github.com/yoojia/go-pipeline/hooks"
	"github.com/yoojia/go-pipeline/http"
	"github.com/yoojia/go-pipeline/kafka"
	"github.com/yoojia/go-pipeline/sql"
	"github.com/yoojia/go-pipeline/util"
	"os"
	"runtime/pprof"
	"time"
)

var (
	BuildTime = "2018.08"
	Version   = "1.0.0"
	GitHash   = "1.0.0"
)

func main() {
	var (
		showHelp      bool   // 显示帮助信息
		showVer       bool   // 显示版本信息
		logConsole    bool   // 显示控制台格式日志
		logTimeFormat string // 设置日志时间格式
		configPath    string // 指定启动配置文件
		genAuthSign   string // 指定密钥，生成认证字符串
		killAfter     string // 自动关闭程序的时间
	)

	var (
		pprofCPUEnabled bool   // 是否开启Cpu Profiling
		pprofCPUOutput  string // 设置CPU Profiling输出文件
	)

	flag.BoolVar(&showHelp, "h", false, "[Help]显示帮助信息")
	flag.BoolVar(&showVer, "v", false, "[Version]显示版本信息")
	flag.StringVar(&logTimeFormat, "f", time.RFC3339, "[Format]日志记录的时间格式")
	flag.StringVar(&configPath, "c", "conf.d", "[Config]指定启动配置文件目录")
	flag.BoolVar(&logConsole, "l", false, "[Log]显示控制台格式的日志信息")
	flag.StringVar(&killAfter, "k", "", "[KillAppAfter]自动关闭程序的时间，如：'10s'")
	// Utils
	flag.StringVar(&genAuthSign, "sign", "", "[GenerateAuthSign]指定密钥，生成认证字符串")
	// DebugConfig configs
	flag.BoolVar(&pprofCPUEnabled, "with-pprof-cpu", false, "[Profiling CPU]是否开启CPU Profiling")
	flag.StringVar(&pprofCPUOutput, "with-pprof-cpu-output", "", "[Profiling CPU Output]CPU Profiling输出文件")

	flag.Parse()
	// Log日志格式化
	zerolog.TimeFieldFormat = logTimeFormat

	// 生成模拟认证数据
	if "" != genAuthSign {
		if signTxt, err := util.GenerateAuthSign(genAuthSign); nil != err {
			panic(err)
		} else {
			fmt.Println(signTxt)
		}
		return
	}

	// 显示程序帮助
	if showHelp {
		showHelpInfo()
		return
	}

	// 显示版本信息
	if showVer {
		showVersionInfo()
		return
	}

	//// 启动主程序
	// PID Support
	pfm := pid.NewPidFileManagerDefault("GoPipeline")
	if err := pfm.Setup(); nil != err {
		fmt.Println(err.Error())
		return
	}
	defer pfm.Cleanup()

	showVersionInfo()

	pipeline := gopl.SharedRouter()

	if logConsole {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	// 自动关闭
	if killDuration := gopl.DurationOrDefault(killAfter, 0); 0 < killDuration {
		log.Info().Msgf("GoPipeline kill after %s", killDuration)
		time.AfterFunc(killDuration, func() {
			pipeline.StopRouter()
		})
	}
	// 调试
	if pprofCPUEnabled {
		if "" == pprofCPUOutput {
			log.Panic().Msgf("Output file is required when CPU Profiling set to enabled")
		}
		f, err := os.Create(pprofCPUOutput)
		if nil != err {
			log.Panic().Err(err).Msgf("Failed to create CPU Profiling output file: %s", pprofCPUOutput)
			return
		}
		defer f.Close()

		log.Info().Msg("### Start with CPU Profiling")
		log.Info().Msgf("### CPU Profiling output: %s", pprofCPUOutput)
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	hook := func(r *gopl.GoPipeline) {
		r.RegisterTerminateHook(hooks.TPSReporterTerminateHook)

	}

	buildIn := func(r *gopl.GoPipeline) {
		// common
		r.AutoRegister(new(common.GoPLMockInput))
		r.AutoRegister(new(common.GoPLDeliverCountInput))
		r.AutoRegister(new(common.GoPLConsoleOutput))
		r.AutoRegister(new(common.GoPLProcMemInfoDecoder))
		r.AutoRegister(new(common.GoPLFilePollingInput))

		// http
		r.RegisterStartupHook(http.ServerStartupHook)
		r.RegisterTerminateHook(http.ServerTerminateHook)
		r.AutoRegister(new(http.GoPLHttpServerInput))
		r.AutoRegister(new(http.GoPLWebSocketClientInput))
		r.AutoRegister(new(http.GoPLWebSocketServerOutput))

		// MQ
		r.AutoRegister(new(kafka.GoPLKafkaProducerOutput))
		// DB
		r.AutoRegister(new(sql.GoPLMySQLQueryInput))
	}

	// 初始化
	pipeline.Prepare(hook, buildIn)
	pipeline.Setup(configPath)
	pipeline.Init()
	pipeline.StartRoute()
}

func showHelpInfo() {
	fmt.Fprintf(os.Stderr, `GoPipeline	yoojiachen@gmail.com
Version:  %s
BuildAt:  %s
GitHash:  %s

Options:

`, Version, BuildTime, GitHash)
	flag.PrintDefaults()
}

func showVersionInfo() {
	fmt.Printf("GoPipeline	build:%s	version:%s	  githash:%s \n", BuildTime, Version, GitHash)
}
