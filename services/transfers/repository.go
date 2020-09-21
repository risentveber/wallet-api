package transfers

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

var ErrNoRowsAffected = errors.New("db_no_rows_affected")

type entityNotFound struct {
	uuid.UUID
}

func (e entityNotFound) Error() string {
	return "entity with uuid: " + e.String() + " not found"
}

func validateAffected(res sql.Result) error {
	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNoRowsAffected
	}

	return nil
}

type InnerTransferCallback func(sender, receiver Account, a InnerTransferActions) error

type InnerTransferActions interface {
	CreateTransfer(transfer Transfer) error
	UpdateBalance(accountID uuid.UUID, diff decimal.Decimal) error
	CreateTransferPart(tp TransferPart) error
}

type Repository interface {
	GetPrecision(ctx context.Context, currency string) (precision uint, exists bool, err error)
	// For separation business logic errors from database errors
	IsTransferIDUsedError(err error) bool
	IsEntityNotFoundError(uuid uuid.UUID, err error) bool
	// locks sender and receiver and manipulates data inside db transaction,
	// return entity not found error if sender or receiver don't exist
	CreateInnerTransferTransactionWithLock(ctx context.Context, sender, receiver uuid.UUID, c InnerTransferCallback) error
	GetAccounts(ctx context.Context, limit uint) ([]Account, error)
	GetTransferInfos(ctx context.Context, accountID uuid.UUID, limit uint) ([]TransferInfo, error)
}

func NewRepository(db *sql.DB) Repository {
	return repository{db}
}

type repository struct {
	db *sql.DB
}

func (r repository) GetPrecision(ctx context.Context, currencyCode string) (uint, bool, error) {
	var precision uint

	row := r.db.QueryRowContext(ctx, `SELECT precision FROM currencies WHERE code=$1;`, currencyCode)
	switch err := row.Scan(&precision); err {
	case sql.ErrNoRows:
		return 0, false, nil
	case nil:
		return precision, true, nil
	default:
		return 0, false, err
	}
}

func (r repository) IsTransferIDUsedError(err error) bool {
	return err != nil && err.Error() == `pq: duplicate key value violates unique constraint "transfers_pkey"`
}

func (r repository) IsEntityNotFoundError(id uuid.UUID, err error) bool {
	if err == nil || id == uuid.Nil {
		return false
	}
	exactErr, ok := err.(entityNotFound)

	return ok && exactErr.UUID == id
}

type innerTransferTxn struct {
	dbTx *sql.Tx
	ctx  context.Context
}

func (tx innerTransferTxn) CreateTransfer(t Transfer) error {
	res, err := tx.dbTx.ExecContext(tx.ctx, `
INSERT INTO transfers(id, type, amount, currency_code)
 VALUES ($1, $2, $3, $4)`, t.ID, t.Type, t.Amount, t.CurrencyCode)
	if err != nil {
		return err
	}

	return validateAffected(res)
}

func (tx innerTransferTxn) UpdateBalance(accountID uuid.UUID, balance decimal.Decimal) error {
	res, err := tx.dbTx.ExecContext(tx.ctx, `
UPDATE accounts
SET balance = $1, updated_at = now()
WHERE id = $2`, balance, accountID)
	if err != nil {
		return err
	}

	return validateAffected(res)
}

func (tx innerTransferTxn) CreateTransferPart(tp TransferPart) error {
	_, err := tx.dbTx.ExecContext(tx.ctx, `
INSERT INTO transfer_parts(transfer_id, account_id, corresponding_account_id, direction)
 VALUES ($1, $2, $3, $4)`, tp.TransferID, tp.AccountID, tp.CorrespondingAccountID, tp.Direction)

	return err
}

func generateFirstEntityNotFoundError(accounts []Account, ids ...uuid.UUID) error {
	for _, id := range ids {
		var found bool
		for _, a := range accounts {
			if a.ID == id {
				found = true

				break
			}
		}
		if !found {
			return entityNotFound{id}
		}
	}

	return nil
}

func (r repository) CreateInnerTransferTransactionWithLock(
	ctx context.Context, sender, receiver uuid.UUID, c InnerTransferCallback) (err error) {
	var rows *sql.Rows
	var tx *sql.Tx

	tx, err = r.db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if err != nil {
			_ = tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	rows, err = tx.QueryContext(ctx, `
		SELECT id, currency_code, balance, created_at, updated_at FROM accounts 
		WHERE id in ($1, $2) ORDER BY 
		CASE
			WHEN id=$1 THEN 1
			ELSE 2
	  	END
		FOR NO KEY UPDATE 
	`, sender, receiver)
	if err != nil {
		return
	}
	defer rows.Close()
	accounts := make([]Account, 0, 2) // expected 2 accounts or lower
	for rows.Next() {
		var a Account
		err = rows.Scan(&a.ID, &a.CurrencyCode, &a.Balance, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return
		}
		accounts = append(accounts, a)
	}

	if err = rows.Err(); err != nil {
		return err
	}
	if len(accounts) < 2 { // nolint gomnd
		err = generateFirstEntityNotFoundError(accounts, sender, receiver)

		return
	}
	err = c(accounts[0], accounts[1], innerTransferTxn{dbTx: tx, ctx: ctx})

	return err
}

func (r repository) GetAccounts(ctx context.Context, limit uint) ([]Account, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, currency_code, balance, created_at, updated_at FROM accounts ORDER BY updated_at LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	accounts := make([]Account, 0, limit)
	var a Account
	for rows.Next() {
		err := rows.Scan(&a.ID, &a.CurrencyCode, &a.Balance, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}

	return accounts, rows.Err()
}

func (r repository) GetTransferInfos(ctx context.Context, accountID uuid.UUID, limit uint) ([]TransferInfo, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT t.id, tp.account_id, tp.corresponding_account_id, t.type, tp.direction, t.currency_code, t.amount, t.created_at
FROM transfer_parts as tp
INNER JOIN transfers as t ON tp.transfer_id = t.id
WHERE tp.account_id = $1 ORDER BY t.created_at DESC
LIMIT $2`, accountID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	transferInfos := make([]TransferInfo, 0, limit)
	var ti TransferInfo
	for rows.Next() {
		err := rows.Scan(
			&ti.ID, &ti.AccountID, &ti.CorrespondingAccountID,
			&ti.Type, &ti.Direction, &ti.CurrencyCode, &ti.Amount, &ti.CreatedAt)
		if err != nil {
			return nil, err
		}
		transferInfos = append(transferInfos, ti)
	}

	return transferInfos, rows.Err()
}
