package main

import (
	"github.com/yoojia/go-pid"
)

func main() {
	pid.NewPidCtrlDefault("GoPipeline").Ctrl()
}
