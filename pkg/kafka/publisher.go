package kafka

import (
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
)

// Publisher handles Kafka message publishing
type Publisher struct {
	producer sarama.SyncProducer
}

// NewPublisher creates a new Kafka publisher
func NewPublisher(brokers []string) (*Publisher, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		producer: producer,
	}, nil
}

// Publish sends a message to Kafka topic
func (p *Publisher) Publish(topic string, value interface{}) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(jsonData),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return err
	}

	log.Printf("Message published to topic %s, partition %d, offset %d", topic, partition, offset)
	return nil
}

// Close closes the producer
func (p *Publisher) Close() error {
	return p.producer.Close()
}

