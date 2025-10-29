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
		brokers := []string{"localhost:9092"}
		publisher, err := kafka.NewPublisher(brokers)
		if err != nil {
			panic(err)
		}
		kafkaPublisher = publisher
	})
	return kafkaPublisher
}

// CloseKafkaPublisher closes Kafka publisher
func CloseKafkaPublisher() {
	if kafkaPublisher != nil {
		kafkaPublisher.Close()
	}
}
