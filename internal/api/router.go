package api

import (
	"database/sql"
	"net/http"
	"payment-gateway/internal/middleware"
	"payment-gateway/internal/repository"
	"payment-gateway/internal/services"

	_ "payment-gateway/docs"

	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/gorilla/mux"
)

func SetupRouter(
	db *sql.DB,
	txnRepo repository.TransactionRepository,
	gatewayRepo repository.GatewayRepository,
	txnService services.TransactionService,
) *mux.Router {
	router := mux.NewRouter()
	router.Use(middleware.LoggingMiddleware)

	// repos which may or may not be shared across routes or route groups

	// v1 routes
	v1 := router.PathPrefix("/api/v1").Subrouter()
	{
		// Txn Routes like deposit, withdrawal
		{
			// Dependencies for txn routes
			txnHandler := NewTransactionHandler(txnService)

			// Payment Route Group
			paymentRoutes := v1.PathPrefix("/payments/{operation}")
			paymentRoutes.Handler(http.HandlerFunc(txnHandler.PaymentHandler)).Methods("POST")
		}

		// Webhooks
		{
			txnHandler := NewTransactionHandler(txnService)

			webhookRoutes := v1.PathPrefix("/webhooks")
			webhookRoutes.Handler(http.HandlerFunc(txnHandler.HandleWebhook)).Methods("POST")
		}

		// Swagger
		{
			v1.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)
		}
	}

	return router
}
