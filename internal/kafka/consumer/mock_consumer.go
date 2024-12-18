package consumer

import (
	"context"

	kafkaGo "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
)

// MockKafkaReader is a mock implementation of kafkaGo.Reader
type MockKafkaReader struct {
	mock.Mock
}

func (m *MockKafkaReader) ReadMessage(ctx context.Context) (kafkaGo.Message, error) {
	args := m.Called(ctx)
	return args.Get(0).(kafkaGo.Message), args.Error(1)
}

func (m *MockKafkaReader) CommitMessages(ctx context.Context, msgs ...kafkaGo.Message) error {
	args := m.Called(ctx, msgs)
	return args.Error(0)
}

func (m *MockKafkaReader) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockKafkaConsumerImpl is a mock implementation of KafkaConsumerImpl
type MockKafkaConsumerImpl struct {
	MockKafkaReader
	batchSize int
}

func NewMockKafkaConsumer(batchSize int) *MockKafkaConsumerImpl {
	return &MockKafkaConsumerImpl{
		MockKafkaReader: MockKafkaReader{},
		batchSize:       batchSize,
	}
}

func (kc *MockKafkaConsumerImpl) GetBatchSize() int {
	return kc.batchSize
}
