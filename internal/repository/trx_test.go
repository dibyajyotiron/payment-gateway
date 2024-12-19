package repository

import (
	"context"
	"payment-gateway/internal/models"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestCreateTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewTransactionRepository(db)

	txn := &models.Transaction{
		Amount:    100,
		Type:      "DEPOSIT",
		Status:    "PENDING",
		GatewayID: 1,
		CountryID: 2,
		UserID:    3,
	}

	mock.ExpectQuery(`INSERT INTO transactions`).
		WithArgs(txn.Amount, txn.Type, txn.Status, txn.GatewayID, txn.CountryID, txn.UserID, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	ctx := context.Background()
	id, err := repo.CreateTransaction(ctx, txn)

	assert.NoError(t, err)
	assert.Equal(t, id, txn.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetTransactions(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewTransactionRepository(db)

	mockRows := sqlmock.NewRows([]string{"id", "amount", "type", "status", "user_id", "gateway_id", "country_id", "created_at"}).
		AddRow(1, 100, "DEPOSIT", "PENDING", 3, 1, 2, time.Now()).
		AddRow(2, 200, "WITHDRAWAL", "COMPLETED", 4, 2, 1, time.Now())

	mock.ExpectQuery(`SELECT id, amount, type, status, user_id, gateway_id, country_id, created_at FROM transactions`).
		WillReturnRows(mockRows)

	transactions, err := repo.GetTransactions()

	assert.NoError(t, err)
	assert.Len(t, transactions, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateTransactionStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewTransactionRepository(db)

	txn := &models.Transaction{
		ID:     1,
		Status: models.PENDING,
	}

	newStatus := models.SUCCESS

	mock.ExpectQuery(`UPDATE transactions SET status = \$1 WHERE id = \$2 returning status`).
		WithArgs(newStatus, txn.ID).
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(newStatus))

	err = repo.UpdateTransactionStatus(txn, newStatus.String())

	assert.NoError(t, err)
	assert.Equal(t, newStatus, txn.Status)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateTransactionsBulk(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewTransactionRepository(db)

	transactions := []*models.Transaction{
		{ID: 1, Status: "COMPLETED"},
		{ID: 2, Status: "FAILED"},
	}

	mock.ExpectExec(`UPDATE transactions SET status = \$1, updated_at = \$2 WHERE id = \$3 AND id = \$4 RETURNING id, status, updated_at`).
		WithArgs("COMPLETED", sqlmock.AnyArg(), 1, "FAILED", sqlmock.AnyArg(), 2).
		WillReturnResult(sqlmock.NewResult(0, 2))

	err = repo.UpdateTransactionsBulk(transactions)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewTransactionRepository(db)

	mockTxn := models.Transaction{
		ID:        1,
		Amount:    100,
		Type:      "DEPOSIT",
		Status:    "PENDING",
		UserID:    3,
		GatewayID: 1,
		CountryID: 2,
		CreatedAt: time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "amount", "type", "status", "user_id", "gateway_id", "country_id", "created_at"}).
		AddRow(mockTxn.ID, mockTxn.Amount, mockTxn.Type, mockTxn.Status, mockTxn.UserID, mockTxn.GatewayID, mockTxn.CountryID, mockTxn.CreatedAt)

	mock.ExpectQuery(`SELECT id, amount, type, status, user_id, gateway_id, country_id, created_at FROM transactions WHERE id = \$1`).
		WithArgs(mockTxn.ID).
		WillReturnRows(rows)

	txn, err := repo.GetTransaction(mockTxn.ID)

	assert.NoError(t, err)
	assert.Equal(t, mockTxn, *txn)

	assert.NoError(t, mock.ExpectationsWereMet())
}
