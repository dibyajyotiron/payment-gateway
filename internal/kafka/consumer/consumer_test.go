package consumer

import (
	"context"
	"errors"
	"testing"

	kafkaGo "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test ReadMessage successfully reads a message
func TestReadMessage(t *testing.T) {
	mockConsumer := NewMockKafkaConsumer(1)

	expectedMessage := kafkaGo.Message{
		Key:   []byte("test-key"),
		Value: []byte("test-value"),
	}

	mockConsumer.On("ReadMessage", mock.Anything).Return(expectedMessage, nil)

	ctx := context.Background()
	msg, err := mockConsumer.ReadMessage(ctx)

	assert.NoError(t, err)
	assert.Equal(t, expectedMessage, msg)
	mockConsumer.AssertExpectations(t)
}

// Test ReadMessage returns an error
func TestReadMessage_Error(t *testing.T) {
	mockConsumer := NewMockKafkaConsumer(1)

	mockConsumer.On("ReadMessage", mock.Anything).Return(kafkaGo.Message{}, errors.New("read error"))

	ctx := context.Background()
	msg, err := mockConsumer.ReadMessage(ctx)

	assert.Error(t, err)
	assert.Empty(t, msg)
	mockConsumer.AssertExpectations(t)
}

// Test CommitMessages successfully commits messages
func TestCommitMessages(t *testing.T) {
	mockConsumer := NewMockKafkaConsumer(1)

	mockConsumer.On("CommitMessages", mock.Anything, mock.Anything).Return(nil)

	ctx := context.Background()
	msg := kafkaGo.Message{Key: []byte("test-key"), Value: []byte("test-value")}
	err := mockConsumer.CommitMessages(ctx, msg)

	assert.NoError(t, err)
	mockConsumer.AssertExpectations(t)
}

// Test CommitMessages returns an error
func TestCommitMessages_Error(t *testing.T) {
	mockConsumer := NewMockKafkaConsumer(1)

	mockConsumer.On("CommitMessages", mock.Anything, mock.Anything).Return(errors.New("commit error"))

	ctx := context.Background()
	msg := kafkaGo.Message{Key: []byte("test-key"), Value: []byte("test-value")}
	err := mockConsumer.CommitMessages(ctx, msg)

	assert.Error(t, err)
	mockConsumer.AssertExpectations(t)
}

// Test Close method closes the consumer correctly
func TestClose(t *testing.T) {
	mockConsumer := NewMockKafkaConsumer(1)

	mockConsumer.On("Close").Return(nil)

	err := mockConsumer.Close()

	assert.NoError(t, err)
	mockConsumer.AssertExpectations(t)
}

// Test Close method returns an error
func TestClose_Error(t *testing.T) {
	mockConsumer := NewMockKafkaConsumer(1)

	mockConsumer.On("Close").Return(errors.New("close error"))

	err := mockConsumer.Close()

	assert.Error(t, err)
	mockConsumer.AssertExpectations(t)
}
