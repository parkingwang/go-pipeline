package gopl

import (
	"github.com/parkingwang/go-conf"
	"io/ioutil"
)

//// 签名 ////

type Trace struct {
	Name      string
	Timestamp int64
}

func (slf Trace) String() string {
	return slf.Name
}

//// 头部 ////

type Headers map[string]string

//// 消息对象 ////

type DataFrame struct {
	traces  []*Trace // 处理流程跟踪列表。此字段按顺序记录所有处理过此消息的插件签名。
	topic   string   // 此消息所属的Topic
	headers Headers  // 消息头部，用以设置额外的参数
	*MultiReader
}

// DataFrame functions

func (slf DataFrame) String() string {
	data := conf.Map{}
	data["traces"] = slf.Traces()
	data["topic"] = slf.Topic()
	data["headers"] = slf.headers
	data["body"], _ = slf.ReadBytes()
	data["body_length"] = slf.BodyLength()
	bytes, _ := MarshalJSON(data)
	return string(bytes)
}

// SetHeaders 设置消息多个Header数值对
func (slf *DataFrame) SetHeaders(headers Headers) {
	if nil == headers || len(headers) == 0 {
		return
	}
	for k, v := range headers {
		slf.SetHeader(k, v)
	}
}

// SetHeader 设置单个Header数值对
func (slf *DataFrame) SetHeader(name string, value string) {
	slf.headers[name] = value
}

// Header 返回指定Name的Header值
func (slf *DataFrame) Header(name string) (string, bool) {
	v, hit := slf.headers[name]
	return v, hit
}

// Header 返回指定Name的Header值。如果Name不存在，返回默认值。
func (slf *DataFrame) HeaderOrDefault(name string, defValue string) string {
	if v, hit := slf.headers[name]; hit {
		return v
	} else {
		return defValue
	}
}

// BodyLength 返回Body的字节数
// 当Body的Reader实现，是可读取长度函数时，GetBodyLength返回其长度值。否则返回-1，表示未知长度。
func (slf *DataFrame) BodyLength() int64 {
	return slf.bodyLength
}

// ReadBytes 返回Body的字节数组。如果读取失败，返回Error
func (slf *DataFrame) ReadBytes() ([]byte, error) {
	return ioutil.ReadAll(slf.GetBody())
}

// Traces 返回消息处理跟踪信息
func (slf *DataFrame) Traces() []*Trace {
	idx := 0
	size := len(slf.traces)
	for i := 0; i < size; i++ {
		idx = i
		if nil == slf.traces[i] {
			break
		}
	}
	return slf.traces[:idx]
}

// Topic 返回消息的Topic
func (slf *DataFrame) Topic() string {
	return slf.topic
}

// Sender 返回消息的Sender插件名称
func (slf *DataFrame) Sender() string {
	tracer := slf.traces[0]
	if nil != tracer {
		return tracer.Name
	} else {
		return "nil-tracer"
	}
}

func (slf *DataFrame) addTrace(pluginName string, timestamp int64) {
	trace := &Trace{
		Name:      pluginName,
		Timestamp: timestamp,
	}
	for !slf.addTrace0(trace) {
		slf.traces = append(slf.traces, make([]*Trace, 2)...)
	}
}

func (slf *DataFrame) setTopic(topic string) {
	slf.topic = topic
}

func (slf *DataFrame) addTrace0(add *Trace) bool {
	for i, t := range slf.traces {
		if nil == t {
			slf.traces[i] = add
			return true
		}
	}
	return false
}
