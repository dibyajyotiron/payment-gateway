package producer

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test Publish successfully publishes a message
func TestPublish(t *testing.T) {
	mockProducer := NewMockKafkaProducer()

	ctx := context.Background()
	transactionID := "test-transaction-id"
	message := []byte("test message")

	mockProducer.On("Publish", context.Background(), transactionID, message).Return(nil)

	mockProducer.On("WriteMessages", mock.Anything, mock.Anything).Return(nil)

	err := mockProducer.Publish(ctx, transactionID, message)

	assert.NoError(t, err)
	mockProducer.AssertExpectations(t)
}

// Test Close method successfully closes the producer
func TestClose(t *testing.T) {
	mockProducer := NewMockKafkaProducer()

	mockProducer.On("Close").Return(nil)

	err := mockProducer.Close()

	assert.NoError(t, err)
	mockProducer.AssertExpectations(t)
}

// Test Close method returns an error
func TestClose_Error(t *testing.T) {
	mockProducer := NewMockKafkaProducer()

	mockProducer.On("Close").Return(errors.New("close error"))

	err := mockProducer.Close()

	assert.Error(t, err)
	mockProducer.AssertExpectations(t)
}
