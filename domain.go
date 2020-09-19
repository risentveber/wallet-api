package wallet_api

import (
	"errors"
	"github.com/shopspring/decimal"
	"time"
)

// Business logic level errors that provide enough information about what went wrong
var (
	ErrAccountsMustBeDifferent = errors.New("accounts must be different")
	ErrAmountMustBePositive    = errors.New("transfer amount must be positive")
	ErrUnsupportedCurrency     = errors.New("currency isn't supported")
	ErrInsufficientFunds       = errors.New("insufficient funds")
	ErrSenderNotExists         = errors.New("sender account not exist")
	ErrReceiverNotExists       = errors.New("receiver account not exist")
	ErrSenderWrongCurrency     = errors.New("sender account wrong currency")
	ErrReceiverWrongCurrency   = errors.New("receiver account wrong currency")
)

// Transfer type enums
const (
	Deposit  = "DEPOSIT"
	Withdraw = "WITHDRAW"
	Internal = "INTERNAL"
)

// Direction enums
const (
	Incoming = "INCOMING"
	Outgoing = "OUTGOING"
)

// Currency model for multiple currencies each one with different precision
type Currency struct {
	Code      string
	Precision uint
}

// Transfer order for system to process
type InnerTransferOrder struct {
	ID                string // uuid4 acts as idempotency key
	SenderAccountID   string
	ReceiverAccountID string
	Amount            decimal.Decimal
	CurrencyCode      string // code
}

type TransferInfo struct {
	ID                     string
	AccountID              string
	CorrespondingAccountID *string
	Type                   string
	Direction              string
	Amount                 decimal.Decimal
	CurrencyCode           string
	CreatedAt              time.Time
}

type Transfer struct {
	ID           string
	Type         string // Deposit, Withdraw, Internal
	Amount       decimal.Decimal
	CurrencyCode string
	CreatedAt    time.Time
}

type TransferPart struct {
	TransferID             string
	AccountID              string
	CorrespondingAccountID *string
	Direction              string
}

// Account representation with time fields that are updated accordingly
type Account struct {
	ID           string
	CurrencyCode string
	Balance      decimal.Decimal
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Business actions
type Service interface {
	CreateTransfer(order InnerTransferOrder) error
	GetTransfersForAccount(accountID string) ([]TransferInfo, error)
	GetAccounts() ([]Account, error)
}
