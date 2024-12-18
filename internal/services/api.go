package services

import (
	"context"
	"net/http"
	"payment-gateway/internal/models"
)

type APIResponseSvc interface {
	StartTransactionProcessing(ctx context.Context, request *models.TransactionRequest) (*models.Transaction, error)
}

type APIResponseSvcImpl struct {
	RespModel *models.APIResponse
}

func NewAPIResponse(dataFormat models.DataFormat) *APIResponseSvcImpl {
	response := &models.APIResponse{
		DataFormat: dataFormat,
	}
	return &APIResponseSvcImpl{RespModel: response}
}

func (a *APIResponseSvcImpl) sendProcessedResponse(w http.ResponseWriter, statusCode int, msg string, data interface{}) {
	a.RespModel.StatusCode = statusCode
	a.RespModel.Message = msg
	a.RespModel.Data = data

	if err := EncodeRequestWithHeader(w, a.RespModel); err != nil {
		// Error handling incase encoding failed
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
		return
	}
}

func (a *APIResponseSvcImpl) NewBadRequestErrorResponse(w http.ResponseWriter, msg string) {
	a.sendProcessedResponse(w, http.StatusBadRequest, msg, nil)
}

// NewStatusOKResponse - GRPC style method, where every status has it's own method
func (a *APIResponseSvcImpl) NewStatusOKResponse(w http.ResponseWriter, msg string, data interface{}) {
	a.sendProcessedResponse(w, http.StatusOK, msg, data)
}

// NewInternalServerErrorResponse - GRPC style method, where every status having it's own method
func (a *APIResponseSvcImpl) NewInternalServerErrorResponse(w http.ResponseWriter, msg string, data interface{}) {
	a.sendProcessedResponse(w, http.StatusInternalServerError, "Internal Server Error", data)
}
