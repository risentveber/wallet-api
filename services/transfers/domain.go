package transfers

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Business logic level errors that provide enough information about what went wrong.
var (
	ErrAccountsMustBeDifferent = errors.New("accounts_must_be_different")
	ErrAmountMustBePositive    = errors.New("transfer_amount_must_be_positive")
	ErrUnsupportedCurrency     = errors.New("currency_not_supported")
	ErrInsufficientFunds       = errors.New("insufficient_funds")
	ErrSenderNotExists         = errors.New("sender_account_not_exist")
	ErrReceiverNotExists       = errors.New("receiver_account_not_exist")
	ErrSenderWrongCurrency     = errors.New("sender_account_wrong_currency")
	ErrReceiverWrongCurrency   = errors.New("receiver_account_wrong_currency")
	ErrEmptyTransferID         = errors.New("transfer_id_is_empty")
	ErrEmptySenderAccountID    = errors.New("sender_account_id_is_empty")
	ErrEmptyReceiverAccountID  = errors.New("receiver_account_id_is_empty")
)

// Transfer type enums.
const (
	Deposit  = "DEPOSIT"
	Withdraw = "WITHDRAW"
	Internal = "INTERNAL"
)

// Direction enums.
const (
	Incoming = "INCOMING"
	Outgoing = "OUTGOING"
)

// Currency model for multiple currencies each one with different precision.
type Currency struct {
	Code      string
	Precision uint
}

// Transfer order for system to process.
type InnerTransferOrder struct {
	ID                uuid.UUID       `json:"id"`
	SenderAccountID   uuid.UUID       `json:"sender_account_id"`
	ReceiverAccountID uuid.UUID       `json:"receiver_account_id"`
	Amount            decimal.Decimal `json:"amount"`
	CurrencyCode      string          `json:"currency_code"`
}

type TransferInfo struct {
	ID                     uuid.UUID       `json:"id"`
	AccountID              uuid.UUID       `json:"account_id"`
	CorrespondingAccountID *uuid.UUID      `json:"corresponding_account_id"`
	Type                   string          `json:"type"`
	Direction              string          `json:"direction"`
	Amount                 decimal.Decimal `json:"amount"`
	CurrencyCode           string          `json:"currency_code"`
	CreatedAt              time.Time       `json:"created_at"`
}

type Transfer struct {
	ID           uuid.UUID
	Type         string // Deposit, Withdraw, Internal
	Amount       decimal.Decimal
	CurrencyCode string
	CreatedAt    time.Time
}

type TransferPart struct {
	TransferID             uuid.UUID
	AccountID              uuid.UUID
	CorrespondingAccountID *uuid.UUID
	Direction              string
}

// Account representation with time fields that are updated accordingly.
type Account struct {
	ID           uuid.UUID
	CurrencyCode string
	Balance      decimal.Decimal
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Business actions.
type Service interface {
	CreateTransfer(ctx context.Context, order InnerTransferOrder) error
	GetTransfersForAccount(ctx context.Context, accountID uuid.UUID) ([]TransferInfo, error)
	GetAccounts(ctx context.Context) ([]Account, error)
}
