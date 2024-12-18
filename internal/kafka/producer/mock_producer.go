package producer

import (
	"context"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
)

// MockKafkaWriter is a mock implementation of kafka.Writer
type MockKafkaWriter struct {
	mock.Mock
}

func (m *MockKafkaWriter) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}

func (m *MockKafkaWriter) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockKafkaProducerImpl is a mock implementation of KafkaProducerImpl
type MockKafkaProducerImpl struct {
	MockKafkaWriter
}

func NewMockKafkaProducer() *MockKafkaProducerImpl {
	return &MockKafkaProducerImpl{
		MockKafkaWriter: MockKafkaWriter{},
	}
}

func (kc *MockKafkaProducerImpl) Publish(ctx context.Context, transactionID string, message []byte) error {
	kc.WriteMessages(ctx) // mock write messages as Publish internally calls this
	args := kc.Called(ctx, transactionID, message)
	return args.Error(0)
}

func (kc *MockKafkaProducerImpl) Close() error {
	args := kc.Called()
	return args.Error(0)
}
