# Inputs 输入组件

## 周期性读取文件输入组件

使用此组件，可以定时周期性地读取一个文件。通常用来读取 `/proc/meminfo` 等系统信息。

**注意：此组件每次读取整个文件的内容作为消息体。** 

**Name:** 

> GoPLFilePollingInput

**Config:**

```toml
[GoPLFilePollingInput]
  disabled = false
  decoder = "GoPLProcMemInfoDecoder"
  topic = "/your-topic"
[GoPLFilePollingInput.InitArgs]
  period = "1s"
  file_path = "/proc/meminfo"
```

- period: 读取文件周期。时间格式为 Golang 的 Duration 格式；
- file_path: 读取的文件绝对路径。
- decoder: 读取文件内容的解码接口；

## Http POST 推送JSON输入组件

使用此组件，绑定一个Http端口，其它客户端可以向此Http地址POST数据。

**Name:**

> GoPLHttpPostInput

**Config:**

```toml
[GoPLHttpPostInput]
  disabled = true
  decoder = "JSONDecoder"
  topic = "/your-topic"
[GoPLHttpPostInput.InitArgs]
  pattern = "/your-uri"
```

客户端POST数据：

> POST host:port/your-uri

**Post data**

客户端POST的数据，Form等表单参数将被忽略，其中Body的数据，将打包到DataFrame对象中。

**注意：** 使用Http服务功能，需要启用HttpServer：

```toml
## 配置Http服务。其它使用Http服务的插件，可以通过注册Handler来处理数据。
[HttpServer]
  disabled = false
  address = ":18880"
```

## 周期性读取发送消息包数量统计输入组件

**Name:**

> GLPLDeliverCountInput

**Config:**

```toml
[GLPLDeliverCountInput]
  disabled = false
  decoder = "JSONDecoder"
  topic = "/your-topic"
```

生成的消息格式：

```json
{
  "type": "fio.count",
  "time": "15:04:05",
  "count.inbounds": 100,
  "count.outbounds": 100,
  "count.filtered": 100,
  "avg.inbounds": 10,
  "avg.outbounds": 10,
  "avg.filtered": 10
}
```
