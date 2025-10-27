package weapon

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

// GetKafkaPublisher returns a singleton Kafka publisher instance
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

// getKafkaBrokers returns Kafka broker addresses from environment
func getKafkaBrokers() []string {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		return []string{"localhost:9092"}
	}
	return []string{brokers} // Simplified for now - can parse comma-separated later
}

// CloseKafkaPublisher closes the Kafka publisher
func CloseKafkaPublisher() error {
	if kafkaPublisher != nil {
		return kafkaPublisher.Close()
	}
	return nil
}

