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

func senderPartFrom(o InnerTransferOrder) TransferPart {
	return TransferPart{
		TransferID:             o.ID,
		AccountID:              o.SenderAccountID,
		CorrespondingAccountID: &o.ReceiverAccountID,
		Direction:              Outgoing,
	}
}

func receiverPartFrom(o InnerTransferOrder) TransferPart {
	return TransferPart{
		TransferID:             o.ID,
		AccountID:              o.ReceiverAccountID,
		CorrespondingAccountID: &o.SenderAccountID,
		Direction:              Incoming,
	}
}

func newActionsInsideTransactionForOrder(o InnerTransferOrder) InnerTransferCallback {
	return func(sender, receiver Account, a InnerTransferActions) error {
		if sender.CurrencyCode != o.CurrencyCode {
			return ErrSenderWrongCurrency
		}
		if receiver.CurrencyCode != o.CurrencyCode {
			return ErrReceiverWrongCurrency
		}
		if sender.Balance.LessThan(o.Amount) {
			return ErrInsufficientFunds
		}
		err := a.CreateTransfer(Transfer{ID: o.ID, Amount: o.Amount, Type: Internal, CurrencyCode: o.CurrencyCode})
		if err != nil {
			return err
		}
		err = a.UpdateBalance(sender.ID, sender.Balance.Add(o.Amount.Neg()))
		if err != nil {
			return err
		}
		err = a.UpdateBalance(receiver.ID, receiver.Balance.Add(o.Amount))
		if err != nil {
			return err
		}
		err = a.CreateTransferPart(senderPartFrom(o))
		if err != nil {
			return err
		}

		return a.CreateTransferPart(receiverPartFrom(o))
	}
}

func (s service) CreateTransfer(ctx context.Context, o InnerTransferOrder) error {
	if o.ID == uuid.Nil {
		return ErrEmptyTransferID
	}
	if o.ReceiverAccountID == uuid.Nil {
		return ErrEmptyReceiverAccountID
	}
	if o.SenderAccountID == uuid.Nil {
		return ErrEmptySenderAccountID
	}
	precision, ok, err := s.repo.GetPrecision(ctx, o.CurrencyCode)
	if err != nil {
		return err
	}
	if !ok {
		return ErrUnsupportedCurrency
	}
	o.Amount = o.Amount.Round(int32(precision))
	if !o.Amount.IsPositive() {
		return ErrAmountMustBePositive
	}
	if o.ReceiverAccountID == o.SenderAccountID {
		return ErrAccountsMustBeDifferent
	}

	err = s.repo.CreateInnerTransferTransactionWithLock(
		ctx, o.SenderAccountID, o.ReceiverAccountID,
		newActionsInsideTransactionForOrder(o),
	)
	if s.repo.IsTransferIDUsedError(err) {
		return nil
	}
	if s.repo.IsEntityNotFoundError(o.SenderAccountID, err) {
		return ErrSenderNotExists
	}
	if s.repo.IsEntityNotFoundError(o.ReceiverAccountID, err) {
		return ErrReceiverNotExists
	}

	return err
}

func (s service) GetTransfersForAccount(ctx context.Context, accountID uuid.UUID) ([]TransferInfo, error) {
	return s.repo.GetTransferInfos(ctx, accountID, 100)
}

func (s service) GetAccounts(ctx context.Context) ([]Account, error) {
	return s.repo.GetAccounts(ctx, 100)
}
