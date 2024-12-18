package producer

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaProducer interface {
	Publish(ctx context.Context, transactionID string, message []byte) error
	Close() error
}

type KafkaProducerImpl struct {
	writer *kafka.Writer
}

// Initialize the Kafka writer
func NewKafkaProducer() *KafkaProducerImpl {
	kafkaURL := os.Getenv("KAFKA_BROKER_URL")
	if kafkaURL == "" {
		kafkaURL = "kafka:9092"
	}

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(kafkaURL),
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
		BatchTimeout:           10 * time.Millisecond,
	}

	log.Println("Kafka writer initialized successfully.")
	return &KafkaProducerImpl{
		writer: writer,
	}
}

// publishes a message to the Kafka topic
func (kc *KafkaProducerImpl) Publish(ctx context.Context, transactionID string, message []byte) error {
	if kc.writer == nil {
		log.Println("Kafka writer is nil, cannot publish to Kafka.")
		return errors.New("kafka writer is not initialized")
	}

	topic := getTopic()

	log.Printf("Publishing message to Kafka topic: %s...", topic)

	kafkaMessage := kafka.Message{
		Key:   []byte(transactionID),
		Value: message,
		Topic: topic,
	}

	err := kc.writer.WriteMessages(ctx, kafkaMessage)
	if err != nil {
		log.Printf("Error publishing to Kafka: %v", err)
		return err
	}

	log.Println("Message successfully published to Kafka on topic " + string(topic))
	return nil
}

// returns the appropriate Kafka topic
func getTopic() string {
	// Ideally we can have the topic name ingested from config
	return "transactions"
}

// Close the writer when the system shut down
func (kc *KafkaProducerImpl) Close() error {
	return kc.writer.Close()
}
