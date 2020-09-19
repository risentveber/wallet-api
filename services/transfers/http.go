package transfers

import (
	"context"
	"encoding/json"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

func NewHTTPHandler(endpoints Endpoints) http.Handler {
	r := mux.NewRouter()
	r.Handle("/transfers/",
		httptransport.NewServer(endpoints.CreateTransfer,
			DecodeCreateTransferRequest, EncodeCreateTransferResponse)).
		Methods("POST")
	r.Handle("/accounts/{account_id}/transfers/",
		httptransport.NewServer(endpoints.GetTransfersForAccount,
			DecodeGetTransfersForAccountRequest, EncodeGetTransfersForAccountResponse)).
		Methods("GET")
	r.Handle("/accounts/",
		httptransport.NewServer(endpoints.GetAccounts, DecodeGetAccountsRequest, EncodeGetAccountsResponse)).
		Methods("GET")

	return r
}

type CommonResponse struct {
	Error   string      `json:"error,omitempty"`
	Result  string      `json:"result"`
	Payload interface{} `json:"payload,omitempty"`
}

func NewCommonResponse(payload interface{}, err error) CommonResponse {
	if err != nil {
		return CommonResponse{Error: err.Error(), Result: "ERROR"}
	}

	return CommonResponse{Payload: payload, Result: "OK"}
}

func DecodeCreateTransferRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req CreateTransferRequest
	err := json.NewDecoder(r.Body).Decode(&req)

	return req, err
}

func EncodeCreateTransferResponse(_ context.Context, w http.ResponseWriter, res interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	response, _ := res.(CreateTransferResponse)

	return json.NewEncoder(w).Encode(NewCommonResponse(nil, response.Err))
}

func DecodeGetTransfersForAccountRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req GetTransfersForAccountRequest
	vars := mux.Vars(r)
	req.AccountID = vars["account_id"]

	return req, nil
}

func EncodeGetTransfersForAccountResponse(_ context.Context, w http.ResponseWriter, res interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	response, _ := res.(GetTransfersForAccountResponse)

	return json.NewEncoder(w).Encode(NewCommonResponse(response.Transfers, response.Err))
}

func DecodeGetAccountsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return GetAccountsRequest{}, nil
}

func EncodeGetAccountsResponse(_ context.Context, w http.ResponseWriter, res interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	response, _ := res.(GetAccountsResponse)

	return json.NewEncoder(w).Encode(NewCommonResponse(response.Accounts, response.Err))
}
