package transfers

import (
	"context"
	"database/sql"
)

type Repository interface {
	GetAccounts(ctx context.Context, limit uint) ([]Account, error)
	GetTransferInfos(ctx context.Context, accountID string, limit uint) ([]TransferInfo, error)
}

func NewRepository(db *sql.DB) Repository {
	return repository{db}
}

type repository struct {
	db *sql.DB
}

func (r repository) GetAccounts(ctx context.Context, limit uint) ([]Account, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, currency_code, balance, created_at, updated_at FROM accounts LIMIT $1
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

func (r repository) GetTransferInfos(ctx context.Context, accountID string, limit uint) ([]TransferInfo, error) {
	rows, err := r.db.QueryContext(ctx, `
SELECT transfers.id, account_id, corresponding_account_id, type, direction, currency_code, amount, created_at
FROM transfer_parts 
INNER JOIN transfers ON transfer_parts.transfer_id = transfers.id
WHERE transfer_parts.account_id = $1 ORDER BY transfers.created_at DESC
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
