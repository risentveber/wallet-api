package transfers

import (
	"context"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func prepare() (Service, sqlmock.Sqlmock, error, func()) {
	db, mock, err := sqlmock.New()
	repo := NewRepository(db)
	service := NewService(repo)
	return service, mock, err, func() {
		db.Close()
	}
}

func newValidOrder() InnerTransferOrder {
	order := InnerTransferOrder{}
	order.ID = uuid.New()
	order.ReceiverAccountID = uuid.New()
	order.SenderAccountID = uuid.New()
	order.CurrencyCode = "USD"
	order.Amount, _ = decimal.NewFromString("14.234")
	return order
}

func mustUUID(id string) uuid.UUID {
	uuid, err := uuid.Parse(id)
	if err != nil {
		panic(err)
	}
	return uuid
}

func TestService_GetAccounts(t *testing.T) {
	a := assert.New(t)
	svc, mock, err, close := prepare()
	a.NoError(err, "mock initialized")
	defer close()

	rows := sqlmock.NewRows([]string{"id", "currency_code", "balance", "created_at", "updated_at"}).
		AddRow("3AA42E32-1117-4533-A1B2-86714E9F842E", "USD", "100", time.Now(), time.Now()).
		AddRow("2A9E457A-641F-4484-BA77-B4F6ED4E6633", "EUR", "100", time.Now(), time.Now())
	mock.ExpectQuery("SELECT").WillReturnRows(rows)
	acc, err := svc.GetAccounts(context.Background())
	a.NoError(err)
	a.Equal(2, len(acc), "both accounts selected")
}

func TestService_GetTransfers(t *testing.T) {
	a := assert.New(t)
	svc, mock, err, close := prepare()
	a.NoError(err, "mock initialized")
	defer close()

	rows := sqlmock.NewRows([]string{
		"t.id", "tp.account_id",
		"tp.corresponding_account_id", "t.type", "tp.direction",
		"t.currency_code", "t.amount", "t.created_at",
	}).
		AddRow("4cf1ba3e-3598-4abc-aa4d-351dcb6fe266", "78c3c61f-70fa-477d-88fe-9767638b61a0",
			"742dda95-3205-49fb-8bad-5cac4de9ee39", "INTERNAL", "OUTGOING",
			"USD", "5.32", time.Now())
	mock.ExpectQuery("SELECT").WillReturnRows(rows)
	id := mustUUID("78c3c61f-70fa-477d-88fe-9767638b61a0")
	transfers, err := svc.GetTransfersForAccount(context.Background(), id)
	a.NoError(err)
	a.Equal(1, len(transfers), "one transfer selected")
}

func TestService_CreateTransfer_validate1(t *testing.T) {
	a := assert.New(t)
	svc, mock, err, close := prepare()
	a.NoError(err, "mock initialized")
	defer close()

	order := InnerTransferOrder{}
	err = svc.CreateTransfer(context.Background(), order)
	a.Equal(ErrEmptyTransferID, err)
	order.ID = uuid.New()
	err = svc.CreateTransfer(context.Background(), order)
	a.Equal(ErrEmptyReceiverAccountID, err)
	order.ReceiverAccountID = uuid.New()
	err = svc.CreateTransfer(context.Background(), order)
	a.Equal(ErrEmptySenderAccountID, err)
	order.SenderAccountID = order.ReceiverAccountID
	err = svc.CreateTransfer(context.Background(), order)
	a.Equal(ErrAccountsMustBeDifferent, err)
	order.SenderAccountID = uuid.New()
	currencyRows := sqlmock.NewRows([]string{"precision"}).
		AddRow("2")
	mock.ExpectQuery("^SELECT precision FROM currencies").WillReturnRows(currencyRows)

	err = svc.CreateTransfer(context.Background(), order)
	a.Equal(ErrAmountMustBePositive, err)
	a.NoError(mock.ExpectationsWereMet())
}

func TestService_CreateTransfer_InsufficientFunds(t *testing.T) {
	a := assert.New(t)
	svc, mock, err, close := prepare()
	a.NoError(err, "mock initialized")
	defer close()
	order := newValidOrder()
	currencyRows := sqlmock.NewRows([]string{"precision"}).
		AddRow("2")
	mock.ExpectQuery("^SELECT precision FROM currencies").WillReturnRows(currencyRows)
	mock.ExpectBegin()
	accountRows := sqlmock.NewRows([]string{"id", "currency_code", "balance", "created_at", "updated_at"}).
		AddRow(order.SenderAccountID, "USD", "10", time.Now(), time.Now()).
		AddRow(order.ReceiverAccountID, "USD", "0", time.Now(), time.Now())
	mock.ExpectQuery("^SELECT (.+) FROM accounts").WillReturnRows(accountRows)
	mock.ExpectRollback()

	err = svc.CreateTransfer(context.Background(), order)
	a.Equal(ErrInsufficientFunds, err)
	a.NoError(mock.ExpectationsWereMet())
}

func TestService_CreateTransfer_WrongSenderCurrency(t *testing.T) {
	a := assert.New(t)
	svc, mock, err, close := prepare()
	a.NoError(err, "mock initialized")
	defer close()
	order := newValidOrder()
	currencyRows := sqlmock.NewRows([]string{"precision"}).
		AddRow("2")
	mock.ExpectQuery("^SELECT precision FROM currencies").WillReturnRows(currencyRows)
	mock.ExpectBegin()
	accountRows := sqlmock.NewRows([]string{"id", "currency_code", "balance", "created_at", "updated_at"}).
		AddRow(order.SenderAccountID, "BTC", "10", time.Now(), time.Now()).
		AddRow(order.ReceiverAccountID, "USD", "0", time.Now(), time.Now())
	mock.ExpectQuery("^SELECT (.+) FROM accounts").WillReturnRows(accountRows)
	mock.ExpectRollback()

	err = svc.CreateTransfer(context.Background(), order)
	a.Equal(ErrSenderWrongCurrency, err)
	a.NoError(mock.ExpectationsWereMet())
}

func TestService_CreateTransfer_WrongReceiverCurrency(t *testing.T) {
	a := assert.New(t)
	svc, mock, err, close := prepare()
	a.NoError(err, "mock initialized")
	defer close()
	order := newValidOrder()
	currencyRows := sqlmock.NewRows([]string{"precision"}).
		AddRow("2")
	mock.ExpectQuery("^SELECT precision FROM currencies").WillReturnRows(currencyRows)
	mock.ExpectBegin()
	accountRows := sqlmock.NewRows([]string{"id", "currency_code", "balance", "created_at", "updated_at"}).
		AddRow(order.SenderAccountID, "USD", "10", time.Now(), time.Now()).
		AddRow(order.ReceiverAccountID, "BTC", "0", time.Now(), time.Now())
	mock.ExpectQuery("^SELECT (.+) FROM accounts").WillReturnRows(accountRows)
	mock.ExpectRollback()

	err = svc.CreateTransfer(context.Background(), order)
	a.Equal(ErrReceiverWrongCurrency, err)
	a.NoError(mock.ExpectationsWereMet())
}

func TestService_CreateTransfer_SenderNotExits(t *testing.T) {
	a := assert.New(t)
	svc, mock, err, close := prepare()
	a.NoError(err, "mock initialized")
	defer close()
	order := newValidOrder()
	currencyRows := sqlmock.NewRows([]string{"precision"}).
		AddRow("2")
	mock.ExpectQuery("^SELECT precision FROM currencies").WillReturnRows(currencyRows)
	mock.ExpectBegin()
	accountRows := sqlmock.NewRows([]string{"id", "currency_code", "balance", "created_at", "updated_at"}).
		AddRow(order.ReceiverAccountID, "USD", "0", time.Now(), time.Now())
	mock.ExpectQuery("^SELECT (.+) FROM accounts").WillReturnRows(accountRows)
	mock.ExpectRollback()

	err = svc.CreateTransfer(context.Background(), order)
	a.Equal(ErrSenderNotExists, err)
	a.NoError(mock.ExpectationsWereMet())
}

func TestService_CreateTransfer_ReceiverNotExits(t *testing.T) {
	a := assert.New(t)
	svc, mock, err, close := prepare()
	a.NoError(err, "mock initialized")
	defer close()
	order := newValidOrder()
	currencyRows := sqlmock.NewRows([]string{"precision"}).
		AddRow("2")
	mock.ExpectQuery("^SELECT precision FROM currencies").WillReturnRows(currencyRows)
	mock.ExpectBegin()
	accountRows := sqlmock.NewRows([]string{"id", "currency_code", "balance", "created_at", "updated_at"}).
		AddRow(order.SenderAccountID, "USD", "10", time.Now(), time.Now())
	mock.ExpectQuery("^SELECT (.+) FROM accounts").WillReturnRows(accountRows)
	mock.ExpectRollback()

	err = svc.CreateTransfer(context.Background(), order)
	a.Equal(ErrReceiverNotExists, err)
	a.NoError(mock.ExpectationsWereMet())
}

func TestService_CreateTransfer_IdempotencyKeyUsed(t *testing.T) {
	a := assert.New(t)
	svc, mock, err, close := prepare()
	a.NoError(err, "mock initialized")
	defer close()
	order := newValidOrder()
	currencyRows := sqlmock.NewRows([]string{"precision"}).
		AddRow("2")
	mock.ExpectQuery("^SELECT precision FROM currencies").WillReturnRows(currencyRows)
	mock.ExpectBegin()
	accountRows := sqlmock.NewRows([]string{"id", "currency_code", "balance", "created_at", "updated_at"}).
		AddRow(order.SenderAccountID, "USD", "20", time.Now(), time.Now()).
		AddRow(order.ReceiverAccountID, "USD", "0", time.Now(), time.Now())
	mock.ExpectQuery("^SELECT (.+) FROM accounts").WillReturnRows(accountRows)
	mock.ExpectExec("INSERT INTO transfers").WillReturnError(
		errors.New(`pq: duplicate key value violates unique constraint "transfers_pkey"`))
	mock.ExpectRollback()

	err = svc.CreateTransfer(context.Background(), order)
	a.NoError(err, "No error if key was already used")
	a.NoError(mock.ExpectationsWereMet())
}

type fakeDriverResult struct {
	affectedCount int64
}

func (f fakeDriverResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (f fakeDriverResult) RowsAffected() (int64, error) {
	return f.affectedCount, nil
}

func newFakeDriverResult(count int64) driver.Result {
	return fakeDriverResult{count}
}

func TestService_CreateTransfer_AllNice(t *testing.T) {
	a := assert.New(t)
	svc, mock, err, close := prepare()
	a.NoError(err, "mock initialized")
	defer close()
	order := newValidOrder()
	currencyRows := sqlmock.NewRows([]string{"precision"}).
		AddRow("2")
	mock.ExpectQuery("^SELECT precision FROM currencies").WillReturnRows(currencyRows)
	mock.ExpectBegin()
	accountRows := sqlmock.NewRows([]string{"id", "currency_code", "balance", "created_at", "updated_at"}).
		AddRow(order.SenderAccountID, "USD", "20", time.Now(), time.Now()).
		AddRow(order.ReceiverAccountID, "USD", "0", time.Now(), time.Now())
	mock.ExpectQuery("^SELECT (.+) FROM accounts").WillReturnRows(accountRows)
	mock.ExpectExec("INSERT INTO transfers").WillReturnResult(newFakeDriverResult(1))
	mock.ExpectExec("UPDATE accounts").WillReturnResult(newFakeDriverResult(1))
	mock.ExpectExec("UPDATE accounts").WillReturnResult(newFakeDriverResult(1))
	mock.ExpectExec("INSERT INTO transfer_parts").WillReturnResult(newFakeDriverResult(1))
	mock.ExpectExec("INSERT INTO transfer_parts").WillReturnResult(newFakeDriverResult(1))
	mock.ExpectCommit()

	err = svc.CreateTransfer(context.Background(), order)
	a.NoError(err, "No error if key was already used")
	a.NoError(mock.ExpectationsWereMet())
}
