package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"payment-gateway/db"
	"payment-gateway/internal/api"
	kafkaConsumer "payment-gateway/internal/kafka/consumer"
	kafkaProducer "payment-gateway/internal/kafka/producer"
	"payment-gateway/internal/repository"
	"payment-gateway/internal/services"
)

func main() {
	// Create a wait group to wait for background tasks to finish
	wg := &sync.WaitGroup{}

	// Set up a context with cancel to handle graceful shutdown
	ctx, cancelFunc := context.WithCancel(context.Background())

	// Initialize the database connection
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	dbURL := "postgres://" + dbUser + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/" + dbName + "?sslmode=disable"

	db.InitializeDB(dbURL)
	defer db.Close() // Ensure that the DB connection is closed on shutdown

	// Initialize Kafka producer
	producer := kafkaProducer.NewKafkaProducer()
	defer func() {
		if err := producer.Close(); err != nil {
			log.Printf("Error closing Kafka producer: %v", err)
		}
	}()

	// Set up repositories
	txnRepo := repository.NewTransactionRepository(db.GetDB())
	gatewayRepo := repository.NewGatewayRepository(db.GetDB())

	// Initialize Kafka consumer
	batchSize := os.Getenv("CONSUMER_BATCH_SIZE")
	consumer := kafkaConsumer.NewKafkaConsumer(batchSize)

	// Create the transaction service
	txnService := services.NewTransactionService(txnRepo, gatewayRepo, producer, consumer)

	// Start consuming Kafka messages in a goroutine

	wg.Add(1)
	go func() {
		txnService.Consume(ctx)
		wg.Done()
	}()

	// Set up the HTTP server and routes
	router := api.SetupRouter(db.GetDB(), txnRepo, gatewayRepo, txnService)

	// Start the HTTP server on port 8080
	server := &http.Server{Addr: ":8080", Handler: router}
	go func() {
		log.Println("Starting HTTP server on port 8080...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	ListenForClosingSignals(cancelFunc)

	// Wait for the Kafka consumer and HTTP server to finish
	wg.Wait()

	// Gracefully shut down the HTTP server
	if err := server.Shutdown(ctx); err != nil && err != http.ErrServerClosed {
		log.Printf("HTTP server shutdown failed: %v", err)
	} else {
		log.Println("HTTP server gracefully shut down.")
	}

}

// ListenForClosingSignals Sets up a channel to listen for termination signals (Ctrl+C, SIGTERM, etc.)
// Blocks until a termination signal is received
// Triggers cancellation to stop background processes
//
//	Blocking function
func ListenForClosingSignals(cancelFunc context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	log.Println("Shutdown signal received, starting graceful shutdown...")

	cancelFunc()
}
