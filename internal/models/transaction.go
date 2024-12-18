package models

import (
	"time"
)

type TransactionStatus string

const INIT TransactionStatus = "INIT"
const KAFKA_PUBLISH_FAILED TransactionStatus = "KAFKA_PUBLISH_FAILED"
const PENDING TransactionStatus = "PENDING"
const SUCCESS TransactionStatus = "SUCCESS"
const FAILED TransactionStatus = "FAILED"

func (t TransactionStatus) String() string {
	return string(t)
}

type TransactionType string

func (t *TransactionType) String() string {
	if t == nil {
		return ""
	}
	return string(*t)
}

const DEPOSIT TransactionType = "DEPOSIT"
const WITHDRAWAL TransactionType = "WITHDRAWAL"

type Transaction struct {
	ID        int               `json:"id" xml:"id"`
	Amount    float64           `json:"amount" xml:"amount"`
	Type      string            `json:"type" xml:"type"`
	Status    TransactionStatus `json:"status" xml:"status"`
	CreatedAt time.Time         `json:"created_at" xml:"created_at"`
	UpdatedAt time.Time         `json:"updated_at" xml:"updated_at"`
	GatewayID int               `json:"gateway_id" xml:"gateway_id"`
	CountryID int               `json:"country_id" xml:"country_id"`
	UserID    int               `json:"user_id" xml:"user_id"`
}

// a standard request structure for the transactions
type TransactionRequest struct {
	UserID     int             `json:"user_id" xml:"user_id"`
	Amount     float64         `json:"amount" xml:"amount"`
	Currency   string          `json:"currency" xml:"currency"`
	CountryID  int             `json:"country_id" xml:"country_id"`
	Type       TransactionType `json:"type" xml:"type"` // "deposit" or "withdrawal"
	DataFormat DataFormat      `json:"-" xml:"-"`
}

type TransactionWebhookResponse struct {
	TxnID      int               `json:"txn_id" xml:"txn_id"`
	Status     TransactionStatus `json:"status" xml:"status"`
	UpdatedAt  time.Time         `json:"updated_at" xml:"updated_at"`
	DataFormat DataFormat        `json:"-" xml:"-"`
}
