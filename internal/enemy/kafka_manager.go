package enemy

import (
	"log"
	"os"
	"sync"

	"network-sec-micro/pkg/kafka"
)

var (
	kafkaPublisher *kafka.Publisher
	kafkaOnce      sync.Once
)

func GetKafkaPublisher() (*kafka.Publisher, error) {
	var err error
	kafkaOnce.Do(func() {
		brokers := getKafkaBrokers()
		log.Printf("Initializing Kafka publisher with brokers: %v", brokers)
		kafkaPublisher, err = kafka.NewPublisher(brokers)
		if err != nil {
			log.Printf("Failed to initialize Kafka publisher: %v", err)
			return
		}
		log.Println("Kafka publisher initialized successfully")
	})
	return kafkaPublisher, err
}

func getKafkaBrokers() []string {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		return []string{"localhost:9092"}
	}
	return []string{brokers}
}

func CloseKafkaPublisher() error {
	if kafkaPublisher != nil {
		return kafkaPublisher.Close()
	}
	return nil
}
