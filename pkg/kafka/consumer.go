package kafka

import (
	"log"
	"sync"

	"github.com/IBM/sarama"
)

// MessageHandler handles incoming Kafka messages
type MessageHandler func(message []byte) error

// Consumer handles Kafka message consumption
type Consumer struct {
	consumer sarama.ConsumerGroup
	handler  MessageHandler
	topics   []string
	wg       sync.WaitGroup
}

// ConsumerGroupHandler implements sarama.ConsumerGroupHandler
type ConsumerGroupHandler struct {
	handler MessageHandler
}

func (h ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			if err := h.handler(message.Value); err != nil {
				log.Printf("Error processing message: %v", err)
				// Don't commit offset on error - message will be reprocessed
				continue
			}

			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(brokers []string, groupID string, topics []string, handler MessageHandler) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRoundRobin()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		consumer: consumer,
		handler:  handler,
		topics:   topics,
	}, nil
}

// Start starts consuming messages
func (c *Consumer) Start() error {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			handler := ConsumerGroupHandler{handler: c.handler}
			err := c.consumer.Consume(nil, c.topics, handler)
			if err != nil {
				log.Printf("Error from consumer: %v", err)
				return
			}
		}
	}()

	return nil
}

// Close closes the consumer
func (c *Consumer) Close() error {
	err := c.consumer.Close()
	c.wg.Wait()
	return err
}

