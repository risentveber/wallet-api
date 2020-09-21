package integration

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/risentveber/wallet-api/services/transfers"
)

const (
	host = "test-api.docker.local:8080"
	base = "http://" + host
)

func integrationTestsDisabled() bool {
	return os.Getenv("INTEGRATION_TEST") == ""
}

func makePost(path string, requestData map[string]string) (int, transfers.CommonResponse, error) {
	result := transfers.CommonResponse{}
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return 0, result, err
	}

	res, err := http.Post(base+path, "application/json", bytes.NewBuffer(requestBody))

	if err != nil {
		return 0, result, err
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, result, err
	}
	err = json.Unmarshal(data, &result)
	return res.StatusCode, result, err
}

func makeGet(path string) (int, transfers.CommonResponse, error) {
	result := transfers.CommonResponse{}
	res, err := http.Get(base + path)

	if err != nil {
		return 0, result, err
	}
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, result, err
	}
	err = json.Unmarshal(data, &result)
	return res.StatusCode, result, err
}

func applyMigrations(logger log.Logger) error {
	db, err := sql.Open("postgres",
		"host=test-postgres.docker.local port=5432 user=postgres password=test dbname=postgres sslmode=disable")
	if err != nil {
		return err
	}
	defer db.Close()
	migrate.SetTable("migrations")
	migrations := &migrate.FileMigrationSource{
		Dir: "../migrations",
	}
	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)

	if err != nil {
		return err
	}
	_ = logger.Log("msg", fmt.Sprintf("Applied %d migrations!", n))

	seedSQL, err := ioutil.ReadFile("./init.sql")
	if err != nil {
		return err
	}
	_, err = db.Exec(string(seedSQL))
	return err
}

var logger = log.NewLogfmtLogger(os.Stdout)

func prepareDB() error {
	dir, _ := os.Getwd()
	logger.Log("msg", "pwd is +"+dir)
	err := Retry(5*time.Second, 10, func() error {
		tcpAddr, err := net.ResolveTCPAddr("tcp", host)
		if err != nil {
			return err
		}
		conn, err := net.DialTCP("tcp", nil, tcpAddr)
		if err != nil {
			return err
		}
		_ = conn.Close()
		return nil
	}, logger)
	if err != nil {
		_ = logger.Log("fatal", err.Error())
		return err
	}
	err = applyMigrations(logger)
	if err != nil {
		_ = logger.Log("fatal", err.Error())
	}
	return err
}

// run sequentially
func TestIntegration(t *testing.T) {
	if integrationTestsDisabled() {
		t.Skipf("skip integration")
	}
	err := prepareDB()
	if err != nil {
		t.Fatalf("preparation fail %s", err.Error())
	}
	t.Run("ListAccounts", testListAccounts)
	t.Run("Ops1", generateCheckTransferCount("1836981E-7BCE-4356-99A5-A001073E51FE", 1, 0, "1000"))
	t.Run("Balance1", generateCheckBalance("1836981E-7BCE-4356-99A5-A001073E51FE", "1000USD"))
	t.Run("Balance2", generateCheckBalance("8FF54AAA-31D7-4A04-908A-6FA375030432", "100USD"))
	t.Run("TransferWrongCurrency",
		generateTransfer(
			"EE795FCB-F656-4E2B-A095-4B872670D6F7",
			"1836981E-7BCE-4356-99A5-A001073E51FE",
			"6D75C6A3-212B-426B-9CCA-991CBAD8A007",
			"1", "USD", "receiver_account_wrong_currency",
		))
	t.Run("TransferInsufficientFunds",
		generateTransfer(
			"EE795FCB-F656-4E2B-A095-4B872670D6F7",
			"1836981E-7BCE-4356-99A5-A001073E51FE",
			"8FF54AAA-31D7-4A04-908A-6FA375030432",
			"1001", "USD", "insufficient_funds",
		))
	t.Run("TransferOK",
		generateTransfer(
			"EE795FCB-F656-4E2B-A095-4B872670D6F7",
			"1836981E-7BCE-4356-99A5-A001073E51FE",
			"8FF54AAA-31D7-4A04-908A-6FA375030432",
			"100.111", "USD", "",
		))
	t.Run("TransferOKRetry",
		generateTransfer(
			"EE795FCB-F656-4E2B-A095-4B872670D6F7",
			"1836981E-7BCE-4356-99A5-A001073E51FE",
			"8FF54AAA-31D7-4A04-908A-6FA375030432",
			"100.111", "USD", "",
		))
	t.Run("BalanceAfterTransfer1",
		generateCheckBalance("1836981E-7BCE-4356-99A5-A001073E51FE", "899.89USD"))
	t.Run("BalanceAfterTransfer2",
		generateCheckBalance("8FF54AAA-31D7-4A04-908A-6FA375030432", "200.11USD"))
	t.Run("OpsAfterTransfer1",
		generateCheckTransferCount("1836981E-7BCE-4356-99A5-A001073E51FE", 1, 1, "1100.11"))
	t.Run("OpsAfterTransfer2",
		generateCheckTransferCount("8FF54AAA-31D7-4A04-908A-6FA375030432", 2, 0, "200.11"))

	// DB in container is cleared outside tests
}

func testListAccounts(t *testing.T) {
	a := assert.New(t)
	status, res, err := makeGet("/accounts/")
	a.NoError(err)
	a.Equal(http.StatusOK, status)
	a.Equal("", res.Error)
	a.Equal("OK", res.Result)
	accounts, _ := res.Payload.([]interface{})
	a.Equal(6, len(accounts), "must return created in init.sql")
}

func generateCheckTransferCount(id string, incoming, outgoing int, totalTurnover string) func(t *testing.T) {
	return func(t *testing.T) {
		a := assert.New(t)
		status, res, err := makeGet("/accounts/" + id + "/transfers/")
		a.NoError(err)
		a.Equal(http.StatusOK, status)
		a.Equal("", res.Error)
		a.Equal("OK", res.Result)
		transfers, _ := res.Payload.([]interface{})
		a.Equal(incoming+outgoing, len(transfers), "must return expected transfer count")
		var actualIncoming, actualOutgoing int
		var sum decimal.Decimal
		for _, t := range transfers {
			props, _ := t.(map[string]interface{})
			if props == nil {
				continue
			}
			direction, _ := props["direction"].(string)
			if direction == "OUTGOING" {
				actualOutgoing++
			}
			if direction == "INCOMING" {
				actualIncoming++
			}
			amount, _ := props["amount"].(string)
			amountDecimal, err := decimal.NewFromString(amount)
			a.NoError(err, "amount is valid decimal")
			sum = sum.Add(amountDecimal)
		}
		a.Equal(incoming, actualIncoming, "must return expected incoming transfers count")
		a.Equal(outgoing, actualOutgoing, "must return expected incoming transfers count")
		a.Equal(totalTurnover, sum.String(), "total turnover as expected")
	}
}

func generateCheckBalance(accountID, balance string) func(t *testing.T) {
	return func(t *testing.T) {
		a := assert.New(t)
		status, res, err := makeGet("/accounts/")
		a.NoError(err)
		a.Equal(http.StatusOK, status)
		a.Equal("", res.Error)
		a.Equal("OK", res.Result)

		accounts, _ := res.Payload.([]interface{})
		balanceByID := make(map[string]string)
		isRightOrder := true
		prevTime := time.Now()
		for _, ac := range accounts {
			props, _ := ac.(map[string]interface{})
			if props == nil {
				continue
			}
			balance, _ := props["balance"].(string)
			currency, _ := props["currency_code"].(string)
			updatedAt, _ := props["updated_at"].(string)
			updatedAtTime, err := time.Parse(time.RFC3339Nano, updatedAt)
			a.NoError(err)
			isRightOrder = isRightOrder && prevTime.After(updatedAtTime)
			id, _ := props["id"].(string)
			balanceByID[id] = balance + currency
		}
		a.True(isRightOrder, "ordered by updated_at DESC")
		a.Equal(balance, balanceByID[strings.ToLower(accountID)])
	}
}

func generateTransfer(transferID, senderID, receiverID, amount, currencyCode, errorCode string) func(t *testing.T) {
	return func(t *testing.T) {
		a := assert.New(t)
		status, res, err := makePost("/transfers/", map[string]string{
			"id":                  transferID,
			"sender_account_id":   senderID,
			"receiver_account_id": receiverID,
			"amount":              amount,
			"currency_code":       currencyCode,
		})
		a.NoError(err)
		a.Equal(http.StatusOK, status)
		if errorCode != "" {
			a.Equal("ERROR", res.Result)
			a.Equal(errorCode, res.Error)
		} else {
			a.Equal("OK", res.Result)
		}
	}
}
