package transfers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

type svcEmptyMock struct{}

func (m svcEmptyMock) CreateTransfer(ctx context.Context, order InnerTransferOrder) error {
	return nil
}
func (m svcEmptyMock) GetTransfersForAccount(ctx context.Context, accountID uuid.UUID) ([]TransferInfo, error) {
	return nil, nil
}
func (m svcEmptyMock) GetAccounts(ctx context.Context) ([]Account, error) {
	return nil, nil
}

var testLogger = log.NewLogfmtLogger(os.Stdout)

func TestGetTrailingSlashRedirect(t *testing.T) {
	a := assert.New(t)
	handler := NewHTTPHandler(NewEndpoints(svcEmptyMock{}), testLogger)
	req, _ := http.NewRequest("GET", "/accounts", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	a.Equal(http.StatusMovedPermanently, response.Code)
	a.Equal("/accounts/", response.Header().Get("location"))
}

func TestGetAccounts(t *testing.T) {
	a := assert.New(t)
	handler := NewHTTPHandler(NewEndpoints(svcEmptyMock{}), testLogger)
	req, _ := http.NewRequest("GET", "/accounts/", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	a.Equal(http.StatusOK, response.Code)
	a.Equal("application/json; charset=utf-8", response.Header().Get("content-type"))
	a.JSONEq(`{"result":"OK", "payload":null}`, response.Body.String())
}

func TestGetTransfersForAccount(t *testing.T) {
	a := assert.New(t)
	handler := NewHTTPHandler(NewEndpoints(svcEmptyMock{}), testLogger)
	req, _ := http.NewRequest("GET", "/accounts/84C7940A-BC65-4B87-A563-E814E520D040/transfers/", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	a.Equal(http.StatusOK, response.Code)
	a.Equal("application/json; charset=utf-8", response.Header().Get("content-type"))
	a.JSONEq(`{"result":"OK", "payload":null}`, response.Body.String())
}

func TestCreateTransferEmptyBody(t *testing.T) {
	a := assert.New(t)
	handler := NewHTTPHandler(NewEndpoints(svcEmptyMock{}), testLogger)
	req, _ := http.NewRequest("POST", "/transfers/", &bytes.Buffer{})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	a.Equal(http.StatusOK, response.Code)
	a.Equal("application/json; charset=utf-8", response.Header().Get("content-type"))
	a.JSONEq(`{"result":"ERROR", "error":"EOF"}`, response.Body.String())
}

func TestCreateTransferValidBody(t *testing.T) {
	a := assert.New(t)
	handler := NewHTTPHandler(NewEndpoints(svcEmptyMock{}), testLogger)
	req, _ := http.NewRequest("POST", "/transfers/",
		bytes.NewBuffer([]byte(`{"id":"AB363360-632B-4643-B93F-0486B764E98D"}`)))
	req.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	a.Equal(http.StatusOK, response.Code)
	a.Equal("application/json; charset=utf-8", response.Header().Get("content-type"))
	a.JSONEq(`{"result":"OK"}`, response.Body.String())
}

type svcMock struct{}

func (m svcMock) CreateTransfer(ctx context.Context, order InnerTransferOrder) error {
	panic(errors.New("panic"))
}

func (m svcMock) GetTransfersForAccount(ctx context.Context, accountID uuid.UUID) ([]TransferInfo, error) {
	id, _ := uuid.Parse("AB363360-632B-4643-B93F-0486B764E98D")
	return []TransferInfo{{
		ID: id, AccountID: id,
		Amount:       decimal.New(1, 1),
		CurrencyCode: "USD", Direction: Incoming, Type: Deposit,
	}}, nil
}

func (m svcMock) GetAccounts(ctx context.Context) ([]Account, error) {
	id, _ := uuid.Parse("AB363360-632B-4643-B93F-0486B764E98D")
	return []Account{{
		ID: id, Balance: decimal.New(1, 1), CurrencyCode: "USD",
	}}, nil
}

func TestResponseFormatGetTransfers(t *testing.T) {
	a := assert.New(t)
	handler := NewHTTPHandler(NewEndpoints(svcMock{}), testLogger)
	req, _ := http.NewRequest("GET", "/accounts/84C7940A-BC65-4B87-A563-E814E520D040/transfers/", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	a.Equal(http.StatusOK, response.Code)
	a.Equal("application/json; charset=utf-8", response.Header().Get("content-type"))
	fmt.Println(response.Body.String())
	a.JSONEq(`{
  "result": "OK",
  "payload": [
    {
      "id": "ab363360-632b-4643-b93f-0486b764e98d",
      "account_id": "ab363360-632b-4643-b93f-0486b764e98d",
      "corresponding_account_id": null,
      "type": "DEPOSIT",
      "direction": "INCOMING",
      "amount": "10",
      "currency_code": "USD",
      "created_at": "0001-01-01T00:00:00Z"
    }
  ]
}`, response.Body.String())
}

func TestResponseFormatGetAccounts(t *testing.T) {
	a := assert.New(t)
	handler := NewHTTPHandler(NewEndpoints(svcMock{}), testLogger)
	req, _ := http.NewRequest("GET", "/accounts/", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	a.Equal(http.StatusOK, response.Code)
	a.Equal("application/json; charset=utf-8", response.Header().Get("content-type"))
	fmt.Println(response.Body.String())
	a.JSONEq(`{
  "result": "OK",
  "payload": [
    {
      "ID": "ab363360-632b-4643-b93f-0486b764e98d",
      "CurrencyCode": "USD",
      "Balance": "10",
      "CreatedAt": "0001-01-01T00:00:00Z",
      "UpdatedAt": "0001-01-01T00:00:00Z"
    }
  ]
}`, response.Body.String())
}
