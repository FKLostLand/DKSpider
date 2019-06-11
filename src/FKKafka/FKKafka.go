package FKKafka

import (
	"FKBase"
	"FKConfig"
	"FKLog"
	"github.com/Shopify/sarama"
	"strings"
	"sync"
)

var (
	err                    error
	producer               sarama.SyncProducer
	syncOnceCreateProducer sync.Once
)

type KafkaSender struct {
	topic string
}

func GetProducer() (sarama.SyncProducer, error) {
	return producer, err
}

//刷新producer
func Refresh() {
	syncOnceCreateProducer.Do(func() {
		conf := sarama.NewConfig()
		conf.Producer.RequiredAcks = sarama.WaitForAll //等待所有备份返回ack
		conf.Producer.Retry.Max = 10                   // 重试次数
		brokerList := FKConfig.CONFIG_KAFKA_BROKERS_STRING
		producer, err = sarama.NewSyncProducer(strings.Split(brokerList, ","), conf)
		if err != nil {
			FKLog.G_Log.Error("Kafka:%v\n", err)
		}
	})
}

func CreateKafkaSender() *KafkaSender {
	return &KafkaSender{}
}

func (p *KafkaSender) SetTopic(topic string) {
	p.topic = topic
}

func (p *KafkaSender) Push(data map[string]interface{}) error {
	val := FKBase.JsonString(data)
	_, _, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.StringEncoder(val),
	})
	return err
}
