package repository

import (
	"context"
	"database/sql"
	"fmt"
	"payment-gateway/internal/models"
	"strings"
	"time"
)

// TransactionRepository defines the interface for transaction-related database operations.
type TransactionRepository interface {
	// CreateTransaction inserts a new transaction into the database.
	//
	//   If insert is successful, `txn.ID` will be populated inside tx and returned
	CreateTransaction(ctx context.Context, txn *models.Transaction) (int, error)
	// UpdateTransactionStatus updates the status of txn in db.
	//
	//   If update is successful, `txn.Status` will be reflecting the provided status
	UpdateTransactionStatus(txn *models.Transaction, status string) error
	UpdateTransactionsBulk(transactions []*models.Transaction) error
	GetTransaction(txnID int) (*models.Transaction, error)
}

// TransactionRepositoryImpl is the concrete implementation of TransactionRepository.
type TransactionRepositoryImpl struct {
	db *sql.DB
}

// NewTransactionRepository creates a new instance of transactionRepository.
func NewTransactionRepository(db *sql.DB) *TransactionRepositoryImpl {
	return &TransactionRepositoryImpl{db: db}
}

// CreateTransaction inserts a new transaction into the database.
//
// Once insert is successful, `txn.ID` will be populated as well as and returned
func (t *TransactionRepositoryImpl) CreateTransaction(context context.Context, txn *models.Transaction) (int, error) {
	query := `INSERT INTO transactions (amount, type, status, gateway_id, country_id, user_id, created_at) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	err := t.db.QueryRow(query, txn.Amount, txn.Type, txn.Status, txn.GatewayID, txn.CountryID, txn.UserID, time.Now()).Scan(&txn.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to insert transaction: %v", err)
	}
	return txn.ID, nil
}

func (t *TransactionRepositoryImpl) GetTransactions() ([]models.Transaction, error) {
	rows, err := t.db.Query(`SELECT id, amount, type, status, user_id, gateway_id, country_id, created_at FROM transactions`)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %v", err)
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var transaction models.Transaction
		if err := rows.Scan(&transaction.ID, &transaction.Amount, &transaction.Type, &transaction.Status, &transaction.UserID, &transaction.GatewayID, &transaction.CountryID, &transaction.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %v", err)
		}
		transactions = append(transactions, transaction)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return transactions, nil
}

func (t *TransactionRepositoryImpl) GetTransaction(txnID int) (*models.Transaction, error) {
	transaction := models.Transaction{}

	err := t.db.QueryRow(`SELECT id, amount, type, status, user_id, gateway_id, country_id, created_at FROM transactions WHERE id = $1`, txnID).Scan(&transaction.ID, &transaction.Amount, &transaction.Type, &transaction.Status, &transaction.UserID, &transaction.GatewayID, &transaction.CountryID, &transaction.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch transactions: %v", err)
	}

	return &transaction, nil
}

// UpdateTransactionStatus updates the status of an existing transaction.
func (t *TransactionRepositoryImpl) UpdateTransactionStatus(txn *models.Transaction, status string) error {
	query := "UPDATE transactions SET status = $1 WHERE id = $2 returning status"
	err := t.db.QueryRow(query, status, txn.ID).Scan(&txn.Status)
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %v", err)
	}
	return nil
}

// UpdateTransactionsBulk updates the status and other columns of multiple transactions based on dynamic filters and columns.
func (t *TransactionRepositoryImpl) UpdateTransactionsBulk(transactions []*models.Transaction) error {
	// Initialize slices for query parts and values
	setClauses := []string{}
	whereClauses := []string{}
	values := []interface{}{}
	i := 1

	// Build the SET clause for the query
	for _, txn := range transactions {
		// Add the ID filter
		whereClauses = append(whereClauses, fmt.Sprintf("id = $%d", i))
		values = append(values, txn.ID)

		// Dynamically append the set clauses for status and updated_at
		setClauses = append(setClauses, fmt.Sprintf("status = $%d, updated_at = $%d", i+1, i+2))
		// Add status and updated_at values to the parameters list
		values = append(values, txn.Status, time.Now())

		i += 2
	}

	// Start building the bulk update query
	query := fmt.Sprintf("UPDATE transactions SET %s WHERE %s RETURNING id, status, updated_at",
		strings.Join(setClauses, ", "), strings.Join(whereClauses, " AND "))

	// Execute the query with the constructed values
	_, err := t.db.Exec(query, values...)
	if err != nil {
		return fmt.Errorf("failed to update transactions: %v", err)
	}

	return nil
}
