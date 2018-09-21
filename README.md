# GoPipeline - 数据管道

GoPipeline是一个被设计用来消息分派和处理的框架。
它用于艾润物联大数据系统的数据源采集前端，将Http WebHook、WebSocket等消息数据源，
通过GoPipeline内部转换数据格式，并分派到Spark、Kafka、WebSocketServer等其它服务。
