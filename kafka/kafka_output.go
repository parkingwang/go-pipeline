package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/parkingwang/go-conf"
	"github.com/rs/zerolog/log"
	"github.com/yoojia/go-pipeline"
	"strconv"
	"strings"
)

//
// Author: 陈哈哈 chenyongjia@parkingwang.com, yoojiachen@gmail.com
//
// Kafka 输出组件
//

type GoPLKafkaProducerOutput struct {
	gopl.AbcSlot

	messageKey   string // Kafka发送数据时的Key
	messageTopic string // Kafka发送消息的Topic
	producer     sarama.AsyncProducer
}

func (slf *GoPLKafkaProducerOutput) Init(args conf.Map) {
	slf.AbcSlot.Init(args)

	brokers, err := args.MustStringArray("brokers")
	if nil != err {
		slf.TagLog(log.Panic).Msg(err.Error())
	}

	if 0 == len(brokers) {
		slf.TagLog(log.Panic).Msg("<brokers> is required")
	}

	if topic, err := args.MustStringNotEmpty("message_topic"); nil != err {
		slf.TagLog(log.Panic).Err(err).Msg(err.Error())
	} else {
		slf.messageTopic = topic
	}

	slf.messageKey = args.MustString("message_key")

	config := sarama.NewConfig()
	config.Producer.Retry.Max = int(args.GetInt64OrDefault("retry_max", 5))

	switch strings.ToLower(args.MustString("required_acks")) {
	case "waitforall":
		config.Producer.RequiredAcks = sarama.WaitForAll
	case "waitforlocal":
		config.Producer.RequiredAcks = sarama.WaitForLocal
	case "noresponse":
		config.Producer.RequiredAcks = sarama.NoResponse
	default:
		config.Producer.RequiredAcks = sarama.NoResponse
	}

	prod, err := sarama.NewAsyncProducer(brokers, config)
	if nil != err {
		slf.TagLog(log.Panic).Err(err).Msgf("Failed to connect to brokers: %s", brokers)
	} else {
		slf.producer = prod
	}
}

func (slf *GoPLKafkaProducerOutput) Output(pack *gopl.DataFrame) {

	topic := pack.HeaderOrDefault("kafka.message.topic", slf.messageTopic)
	key := pack.HeaderOrDefault("kafka.message.key", slf.messageKey)

	partition, err := strconv.Atoi(pack.HeaderOrDefault("kafka.message.partition", "0"))
	if nil != err {
		slf.TagLog(log.Error).Err(err).Msg("Invalid header: kafka.message.partition")
	}

	if bytes, err := pack.ReadBytes(); nil != err {
		slf.TagLog(log.Error).Err(err).Msg("Failed to read pack")
	} else {
		msg := &sarama.ProducerMessage{
			Topic:     topic,
			Key:       sarama.StringEncoder(key),
			Value:     sarama.ByteEncoder(bytes),
			Partition: int32(partition),
		}
		select {
		case slf.producer.Input() <- msg:
			// nop
		case err := <-slf.producer.Errors():
			slf.TagLog(log.Error).Err(err).Msg("Failed to send message to broker")
		}
	}

}

func (slf *GoPLKafkaProducerOutput) Shutdown() {
	if nil != slf.producer {
		if err := slf.producer.Close(); nil != err {
			slf.TagLog(log.Error).Err(err).Msg("Failed to close producer")
		}
	}
}
