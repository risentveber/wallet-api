package transfers

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
)

type CreateTransferRequest struct {
	Order InnerTransferOrder
}

type CreateTransferResponse struct {
	Err error
}

func MakeCreateTransferEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CreateTransferRequest)
		err := s.CreateTransfer(ctx, req.Order)

		return CreateTransferResponse{Err: err}, nil
	}
}

type GetTransfersForAccountRequest struct {
	AccountID uuid.UUID
}

type GetTransfersForAccountResponse struct {
	Transfers []TransferInfo
	Err       error
}

func MakeGetTransfersForAccountEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetTransfersForAccountRequest)
		transfers, err := s.GetTransfersForAccount(ctx, req.AccountID)

		return GetTransfersForAccountResponse{Transfers: transfers, Err: err}, nil
	}
}

type GetAccountsRequest struct{}

type GetAccountsResponse struct {
	Accounts []Account
	Err      error
}

func MakeGetAccountsEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		_ = request.(GetAccountsRequest)
		accounts, err := s.GetAccounts(ctx)

		return GetAccountsResponse{Accounts: accounts, Err: err}, nil
	}
}

func NewEndpoints(s Service) Endpoints {
	return Endpoints{
		CreateTransfer:         MakeCreateTransferEndpoint(s),
		GetAccounts:            MakeGetAccountsEndpoint(s),
		GetTransfersForAccount: MakeGetTransfersForAccountEndpoint(s),
	}
}

type Endpoints struct {
	CreateTransfer         endpoint.Endpoint
	GetTransfersForAccount endpoint.Endpoint
	GetAccounts            endpoint.Endpoint
}
