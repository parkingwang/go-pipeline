package abc

//
// Author: 陈永佳 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//

type AbcShutdown struct {
	sigShutdown chan struct{}
	sigRunning  chan struct{}
}

// Shutdown 调用时，会解除所有协程等待 ShutdownChan() 的阻塞状态，并阻塞等待终止运行信号。
func (slf *AbcShutdown) Shutdown() {
	close(slf.sigShutdown)
	<-slf.sigRunning
}

// ShutdownNow 立即发出终止运行信号，解除 ShutdownChan() 阻塞状态。
func (slf *AbcShutdown) ShutdownNow() {
	slf.SetTerminated()
	slf.Shutdown()
}

// Init 用于初始化状态变量
func (slf *AbcShutdown) Init() {
	slf.sigShutdown = make(chan struct{})
	slf.sigRunning = make(chan struct{})
}

// ShutdownChan 返回一个阻塞状态通道，可以通过Select此通道来判断终止状态。
// 阻塞状态通过 Shutdown 函数解除。
func (slf *AbcShutdown) ShutdownChan() <-chan struct{} {
	return slf.sigShutdown
}

// 发出终止运行信号
func (slf *AbcShutdown) SetTerminated() {
	close(slf.sigRunning)
}

// 判断是否已发送停止信号
func (slf *AbcShutdown) IsStopped() bool {
	select {
	case <-slf.ShutdownChan():
		return true
	default:
		return false
	}
}
