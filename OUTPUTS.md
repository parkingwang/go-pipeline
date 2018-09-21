# 输出组件

## GoPLKafkaProducerOutput - Kafka 生产者输出组件

GoPLKafkaProducerOutput 作为Kafka的Producer，它可以将消息输出到Kafka集群。

### 配置

```toml
# Kafka 输出组件
[GoPLKafkaProducerOutput]
  disabled = false
  topic = "*"
[GoPLKafkaProducerOutput.InitArgs]
  message_key = "test-data"
  message_topic = "go-goplline-test"
  retry_max = 10
  brokers = [
    "node-imac:9092",
    "node-thinkpad:9092",
  ]
```

### 动态配置

通过配置消息DataFrame的Header参数，可以动态地指定每个消息的Kafka参数。这些参数包括：

1. `kafka.message.topic` 消息Topic
1. `kafka.message.key` 消息Key
1. `kafka.message.partition` 消息Partition，默认为0