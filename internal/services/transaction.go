package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	kafkaConsumer "payment-gateway/internal/kafka/consumer"
	kafkaProducer "payment-gateway/internal/kafka/producer"
	"payment-gateway/internal/models"
	"payment-gateway/internal/repository"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

type TransactionService interface {
	StartTransactionProcessing(ctx context.Context, request *models.TransactionRequest) (*models.Transaction, error)
	StartWebhookProcessing(ctx context.Context, request *models.TransactionWebhookResponse) (*models.Transaction, error)
	Consume(ctx context.Context)
}

type TransactionServiceImpl struct {
	cipherSecret      string
	txnRepository     repository.TransactionRepository
	gatewayRepository repository.GatewayRepository
	publisher         kafkaProducer.KafkaProducer
	consumer          kafkaConsumer.KafkaConsumer
}

// ParseTransactionType converts a string to TransactionType with validation
//
//	It will match only lowercased operation values
func ParseTransactionType(operation string) (models.TransactionType, error) {
	switch strings.ToLower(operation) {
	case "deposit":
		return models.DEPOSIT, nil
	case "withdrawal":
		return models.WITHDRAWAL, nil
	default:
		return "", errors.New("invalid operation. Only 'deposit' or 'withdrawal' are allowed")
	}
}

// createTransactionMessage creates transaction as json string
func createTransactionMessage(transaction *models.Transaction) []byte {
	message := map[string]interface{}{
		"id":         transaction.ID,
		"amount":     transaction.Amount,
		"type":       transaction.Type,
		"status":     transaction.Status,
		"createdAt":  transaction.CreatedAt,
		"gateway_id": transaction.GatewayID,
		"country_id": transaction.CountryID,
		"user_id":    transaction.UserID,
	}
	messageBytes, _ := json.Marshal(message)
	return messageBytes
}

func NewTransactionService(txnRepo repository.TransactionRepository, gatewayRepo repository.GatewayRepository, pub kafkaProducer.KafkaProducer, consumer kafkaConsumer.KafkaConsumer) *TransactionServiceImpl {
	return &TransactionServiceImpl{
		txnRepository:     txnRepo,
		gatewayRepository: gatewayRepo,
		publisher:         pub,
		consumer:          consumer,
		// Ideally in prod, this should be injected from config service.
		cipherSecret: os.Getenv("CIPHER_SECRET"),
	}
}

func (t *TransactionServiceImpl) StartTransactionProcessing(ctx context.Context, request *models.TransactionRequest) (*models.Transaction, error) {

	gateways, err := t.gatewayRepository.GetGatewaysByCountryAndCurrency(
		fmt.Sprint(request.CountryID),
		request.Currency,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch gateway from currency, err: %+v", err)
	}

	if len(gateways) < 1 {
		return nil, fmt.Errorf("no Gateway exists for currency: %s, countryID:%d ", request.Currency, request.CountryID)
	}

	// In case we had any priority field,
	// here we could implement that check to find gateway
	// Here, we're picking the latest created gateway
	// we can have our own priority check here
	gateway := gateways[0]

	// c := NewCipherText(s.cipherSecret)
	// maskedUser, err := c.MaskData([]byte(fmt.Sprint(request.UserID)))
	// if err != nil {
	// 	log.Print("Error masking user, message sending to kafka will fail, err: %+v", err)
	// }

	// maskedGateway, err := c.MaskData([]byte(fmt.Sprint(gateway.ID)))
	// if err != nil {
	// 	log.Print("Error masking user, message sending to kafka will fail, err: %+v", err)
	// }

	// Create transaction
	transaction := &models.Transaction{
		UserID:    request.UserID,
		Amount:    request.Amount,
		Type:      request.Type.String(),
		Status:    models.INIT,
		GatewayID: gateway.ID,
		CreatedAt: time.Now(),
		CountryID: request.CountryID,
	}

	// Save to database with retry
	err = RetryOperation(func() error {
		_, err := t.txnRepository.CreateTransaction(ctx, transaction)
		if err != nil {
			return err
		}
		return nil
	}, 5)

	if err != nil {
		log.Printf("Error while creating trx, err: %+v, transaction: %+v", err, *transaction)
	}

	// Publish transaction to the actual gateway
	// Since there is no actual gateway for this
	// task, just leaving this comment here.
	// We might be using `RetryOperation(someFunc, 2)` to publish
	// to the real gateway server

	// If the message is processed successfully by the gateway,
	// mark the status PENDING, as that will help us
	// determine, messages that were not sent to Gateway
	// to run clean up jobs, query dlq etc
	err = RetryOperation(func() error {
		err = t.txnRepository.UpdateTransactionStatus(transaction, models.PENDING.String())
		if err != nil {
			return err
		}
		return nil
	}, 5)

	if err != nil {
		return nil, fmt.Errorf("failed to update transaction status even though gateway already acked the message, err: %+v", err)
	}

	return transaction, nil
}

func (t *TransactionServiceImpl) StartWebhookProcessing(ctx context.Context, request *models.TransactionWebhookResponse) (*models.Transaction, error) {
	// Find the transaction
	txn, err := t.txnRepository.GetTransaction(request.TxnID)
	if err != nil {
		return nil, err
	}

	txn.Status = request.Status
	txn.UpdatedAt = request.UpdatedAt

	// Publish transaction to Kafka using circuit breaker
	message := createTransactionMessage(txn)

	err = PublishWithCircuitBreaker(func() error {
		return t.publisher.Publish(ctx, fmt.Sprint(txn.ID), message)
	})

	if err != nil {
		log.Printf("Error while publishing to Kafka using circuit breaker, err: %+v, msg: %+v", err, message)
		// If publish to kafka fails, change the status so some cron job can pick these and process for refund if eligible
		err = t.txnRepository.UpdateTransactionStatus(txn, models.KAFKA_PUBLISH_FAILED.String())
		if err != nil {
			// If status setting also fails, log it, and have PD alert active.
			return nil, fmt.Errorf("couldn't update status to %s even though kafka publish failed, msg: %+v, err: %+v", models.KAFKA_PUBLISH_FAILED.String(), message, err)
		}
		return nil, fmt.Errorf("failed to publish transaction to Kafka, err: %+v", err)
	}

	return txn, nil
}

func (t *TransactionServiceImpl) Consume(ctx context.Context) {
	var batch []kafka.Message

	for {
		select {
		case <-ctx.Done():
			log.Printf("Server stop signal received, new messages won't be processed")
			return
		default:
		}
		m, err := t.consumer.ReadMessage(context.Background())

		if err != nil {
			log.Printf("Error reading message: %v, message: %+v", err, m)
			continue
		}

		batch = append(batch, m)

		// Commit the batch periodically after every n messages
		if len(batch) >= t.consumer.GetBatchSize() {
			err := t.BulkProcessMessages(batch)
			if err != nil {
				log.Printf("Failed to update webhook messages in db: %v", err)
				return
			}

			if err := t.consumer.CommitMessages(ctx, batch...); err != nil {
				log.Printf("Failed to commit batch: %v", err)
			} else {
				log.Printf("Successfully committed batch of %d messages", len(batch))
			}
			batch = nil // reset batch
		}
	}
}

// BulkProcessMessages will unmarshal kafka messages as transaction
//
//		SInce, webhook will onyl send ID, status and updatedAt, using id,
//	 other two fields will be updated, everything else will remain same
func (t *TransactionServiceImpl) BulkProcessMessages(msgs []kafka.Message) error {
	var transactions []*models.Transaction

	for _, m := range msgs {
		var transaction *models.Transaction

		// Unmarshal Kafka message value into the Transaction struct
		err := json.Unmarshal(m.Value, &transaction)
		if err != nil {
			log.Printf("Failed to unmarshal Kafka message to Transaction: %v", err)
			return err
		}

		log.Printf("Successfully converted Kafka message to Transaction: %+v", transaction)
		transactions = append(transactions, transaction)
	}

	err := t.txnRepository.UpdateTransactionsBulk(transactions)
	if err != nil {
		return err
	}

	return nil
}
