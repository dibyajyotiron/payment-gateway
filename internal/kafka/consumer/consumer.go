package consumer

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	kafkaGo "github.com/segmentio/kafka-go"
)

type KafkaConsumer interface {
	ReadMessage(ctx context.Context) (kafkaGo.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafkaGo.Message) error
	GetBatchSize() int
	Close() error
}

type KafkaConsumerImpl struct {
	reader    *kafkaGo.Reader
	batchSize int
}

// Initialize the Kafka consumer
//
//	If batchSize is not provided or not a valid number, it will default to 1
func NewKafkaConsumer(batchSize string) *KafkaConsumerImpl {
	kafkaURL := os.Getenv("KAFKA_BROKER_URL")
	if kafkaURL == "" {
		kafkaURL = "kafka:9092"
	}

	reader := kafkaGo.NewReader(kafkaGo.ReaderConfig{
		Brokers:     []string{kafkaURL},
		Topic:       getTopic(),
		GroupID:     "transaction-consumer-group",
		StartOffset: kafkaGo.FirstOffset,
		MaxWait:     1 * time.Second,
	})

	log.Println("Kafka reader initialized successfully.")

	batchSizeInt, err := strconv.Atoi(batchSize)
	if err != nil {
		batchSizeInt = 1
	}
	return &KafkaConsumerImpl{reader: reader, batchSize: batchSizeInt}
}

// Consume listens for messages on the Kafka topic
func (kc *KafkaConsumerImpl) ReadMessage(ctx context.Context) (kafkaGo.Message, error) {
	return kc.reader.ReadMessage(ctx)
}

func (kc *KafkaConsumerImpl) GetBatchSize() int {
	return kc.batchSize
}

// CommitMessage commits the message on the particular topic
func (kc *KafkaConsumerImpl) CommitMessages(ctx context.Context, msgs ...kafkaGo.Message) error {
	return kc.reader.CommitMessages(ctx, msgs...)
}

// returns the appropriate Kafka topic
func getTopic() string {
	// Ideally we can have the topic name ingested from config
	return "transactions"
}

// Close the writer when the system shut down
func (kc *KafkaConsumerImpl) Close() error {
	return kc.reader.Close()
}
