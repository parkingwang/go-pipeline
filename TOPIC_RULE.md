# TOPIC匹配规则说明

## TOPIC的作用

topic在配置文件中定义，用于声明消息的路由匹配规则，应用于Input \ Filter 和 Output 插件的消息匹配上。
其中：

- Input 使用它来生产消息源的Topic，它将被直接设置到消息体中。
- Filter/Output 使用它来过滤消息，符合条件的消息才会被处理。

## TOPIC的格式

1. `*` 表示匹配所有消息；
2. `<URL>` 使用URL格式来匹配消息；

Topic使用HTTP URL的格式来配置和解析。使用到URL的规则如下：

> [scheme://]host/path?query

## TOPIC匹配规则

URL的完整Path是Topic匹配的首要条件。如 "/pipe/stats/traffics" 只能匹配到完全相同的Topic。
也可以使用Query来增加对DataFrame消息对象内部Headers的匹配。

例如：

> topic = "/pipe/stats/traffics?Origin=GoPLStatsInput"

它表示Filter或Output接受以"/pipe/status/traffics"为Topic的消息，并且要求消息体Header的`Origin`字段为“GoPLStatsInput”。
其中的"GoPLStatsInput"是插件名。

GoGoPLline框架会为Input/Filter发送消息时自动填充当前插件名称。

消息体的内部字段，见另一说明文档。