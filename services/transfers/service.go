package transfers

import (
	"context"

	"github.com/google/uuid"
)

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return service{repo}
}

func (s service) CreateTransfer(ctx context.Context, o InnerTransferOrder) error {
	if !o.Amount.IsPositive() {
		return ErrAmountMustBePositive
	}
	if o.ReceiverAccountID == o.SenderAccountID {
		return ErrAccountsMustBeDifferent
	}

	return ErrNotImplemented
}

func (s service) GetTransfersForAccount(ctx context.Context, accountID uuid.UUID) ([]TransferInfo, error) {
	return s.repo.GetTransferInfos(ctx, accountID, 100)
}

func (s service) GetAccounts(ctx context.Context) ([]Account, error) {
	return s.repo.GetAccounts(ctx, 100)
}
