package api

import (
	"context"
	"log"
	"net/http"
	"payment-gateway/internal/models"
	"payment-gateway/internal/services"
	"time"

	"github.com/gorilla/mux"
)

type TransactionHandler struct {
	txService services.TransactionService
}

func NewTransactionHandler(service services.TransactionService) *TransactionHandler {
	return &TransactionHandler{txService: service}
}

// PaymentHandler processes deposit or withdrawal transactions.
//
// @Summary      Create a deposit or withdrawal transaction
// @Description  Initializes a transaction in a pending state for either deposit or withdrawal.
// @Tags         Payments
// @Accept       json
// @Accept       xml
// @Produce      json
// @Produce      xml
// @Param        operation    path      string                        true  "Transaction type: 'DEPOSIT' or 'WITHDRAWAL'"  Enums(DEPOSIT, WITHDRAWAL)
// @Param        request      body      models.TransactionRequest     true   "Transaction request payload"
// @Success      200          {object}  models.SuccessAPIResponse            "Transaction processing initialized successfully"
// @Failure      400          {object}  models.BadRequestAPIResponse         "Invalid request body or operation"
// @Failure      500          {object}  models.InternalErrorAPIResponse      "Internal server error"
// @Router       /api/v1/payments/{operation} [post]
func (t *TransactionHandler) PaymentHandler(w http.ResponseWriter, r *http.Request) {
	req := models.TransactionRequest{}

	// Decode request
	if err := services.DecodeRequest(r, &req); err != nil {
		log.Printf("Error decoding req body during deposit, error: %+v", err)
		services.NewAPIResponse(req.DataFormat).NewBadRequestErrorResponse(w, "Invalid request body")
		return
	}

	// Process the path to know if it's deposit/withdrawal
	vars := mux.Vars(r)
	txnType, err := services.ParseTransactionType(vars["operation"])
	if err != nil {
		services.NewAPIResponse(req.DataFormat).NewBadRequestErrorResponse(w, "Invalid operation. Only 'deposit' or 'withdrawal' are allowed.")
		return
	}

	req.Type = txnType

	// Process transaction
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	transaction, err := t.txService.StartTransactionProcessing(ctx, &req)
	if err != nil {
		log.Printf("Error while starting %s processing, req: %+v, error: %+v", txnType.String(), req, err)
		services.NewAPIResponse(req.DataFormat).NewBadRequestErrorResponse(w, "Invalid request body")
		return
	}

	// Respond with success
	services.NewAPIResponse(req.DataFormat).NewStatusOKResponse(w, txnType.String()+" processing initialized", transaction)
}

// HandleWebhook processes incoming webhook updates from the gateway.
//
// @Summary      Process webhook updates
// @Description  Processes webhook responses to update the status of transactions based on the gateway's response.
// @Tags         Webhooks
// @Accept       json
// @Accept       xml
// @Produce      json
// @Produce      xml
// @Param        request      body      models.TransactionWebhookResponse  true   "Webhook response payload"
// @Success      200          {object}  models.APIResponse               "Webhook processing completed successfully"
// @Failure      400          {object}  models.APIResponse               "Invalid request body"
// @Failure      500          {object}  models.APIResponse               "Internal server error"
// @Router       /api/v1/webhooks [post]
func (t *TransactionHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	req := models.TransactionWebhookResponse{
		DataFormat: services.GetDataFormat(r),
	}

	// Decode request
	if err := services.DecodeRequest(r, &req); err != nil {
		log.Printf("Error decoding req body during deposit, error: %+v", err)
		services.NewAPIResponse(req.DataFormat).NewBadRequestErrorResponse(w, "Invalid request body")
		return
	}

	_, err := t.txService.StartWebhookProcessing(context.Background(), &req)
	// In my previous experience, gateway providers mostly care about status code of webhooks
	// to know if they should retry the webhook or not, so ignored the response and only used error
	if err != nil {
		log.Printf("Error Starting webhook processing, error: %+v", err)
		services.NewAPIResponse(req.DataFormat).NewBadRequestErrorResponse(w, "")
		return
	}

	services.NewAPIResponse(req.DataFormat).NewStatusOKResponse(w, "", nil)
}
