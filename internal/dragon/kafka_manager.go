package dragon

import (
	"sync"

	"network-sec-micro/pkg/kafka"
)

var (
	kafkaPublisher *kafka.Publisher
	publisherOnce  sync.Once
)

// GetKafkaPublisher returns singleton Kafka publisher
func GetKafkaPublisher() *kafka.Publisher {
	publisherOnce.Do(func() {
		kafkaPublisher = kafka.NewPublisher()
	})
	return kafkaPublisher
}

// CloseKafkaPublisher closes Kafka publisher
func CloseKafkaPublisher() {
	if kafkaPublisher != nil {
		kafkaPublisher.Close()
	}
}
