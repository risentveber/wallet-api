package transfers

import (
	"context"
	"errors"
)

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return service{repo}
}

func (s service) CreateTransfer(ctx context.Context, order InnerTransferOrder) error {
	return errors.New("not implemented")
}

func (s service) GetTransfersForAccount(ctx context.Context, accountID string) ([]TransferInfo, error) {
	return s.repo.GetTransferInfos(ctx, accountID, 100)
}

func (s service) GetAccounts(ctx context.Context) ([]Account, error) {
	return s.repo.GetAccounts(ctx, 100)
}
