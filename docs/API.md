# API methods

For examples here https://httpie.org/ is used.

## Common response format

All json responses are with `snake_case` field names format.
All errors hidden from HTTP level - you supposed to check result yourself.

```
entity common_response {
  result  string // enum 'OK'|'ERROR'
  error   string // empty if ResultCode is 'OK'
  payload object|array // context-dependent
}
```

## CreateInnerTransfer

`POST <endpoint>/transfers/`

Create internal transfer between the two wallet accounts.
It's important to understand if you run several times with the same `id`
only first succeeded will be really applied, for the rest 'OK' will be returned
without real transfer even if you change other fields.
```
entity inner_transfer_order {
	id                  string // acts as idempotency key
	sender_account_id   string
	reciever_account_id string
	amount              decimal
	currency_code       string 
}
```

Business-level error codes:
- `accounts_must_be_different`
- `transfer_amount_must_be_positive`
- `currency_not_supported`
- `insufficient_funds`
- `sender_account_not_exist`
- `receiver_account_not_exist`
- `sender_account_wrong_currency`
- `receiver_account_wrong_currency`
- `transfer_id_is_empty`
- `sender_account_id_is_empty`
- `receiver_account_id_is_empty`

Example with error:
```
POST /transfers/ HTTP/1.1
Accept: application/json, */*;q=0.5
Accept-Encoding: gzip, deflate
Connection: keep-alive
Content-Length: 214
Content-Type: application/json
Host: localhost:8080
User-Agent: HTTPie/2.2.0

{
    "amount": "1000.0",
    "currency_code": "EUR",
    "id": "ca5bb6ce-1155-4bdb-953f-7267c9bfd82f",
    "receiver_account_id": "8ff54aaa-31d7-4a04-908a-6fa375030432",
    "sender_account_id": "1836981e-7bce-4356-99a5-a001073e51fe"
}

HTTP/1.1 200 OK
Content-Length: 59
Content-Type: application/json; charset=utf-8
Date: Mon, 21 Sep 2020 11:20:03 GMT

{
    "error": "sender_account_wrong_currency",
    "result": "ERROR"
}
```

Example with success
```
POST /transfers/ HTTP/1.1
Accept: application/json, */*;q=0.5
Accept-Encoding: gzip, deflate
Connection: keep-alive
Content-Length: 211
Content-Type: application/json
Host: localhost:8080
User-Agent: HTTPie/2.2.0

{
    "amount": "100",
    "currency_code": "USD",
    "id": "ca5bb6ce-1155-4bdb-953f-7267c9bfd82f",
    "receiver_account_id": "8ff54aaa-31d7-4a04-908a-6fa375030432",
    "sender_account_id": "1836981e-7bce-4356-99a5-a001073e51fe"
}

HTTP/1.1 200 OK
Content-Length: 16
Content-Type: application/json; charset=utf-8
Date: Mon, 21 Sep 2020 11:22:28 GMT

{
    "result": "OK"
}
```
 
### GetPaymentsByAccountID

`GET <endpoint>/accounts/{accountID}/transfers/`

Method returns array transfers ordered by `UpdatedAt` if the count greater than 100, limit to it 100.
Returns empty array when account not found.
```
entity transfer {
    id                       string
    currency_code            string
    amount                   decimal
    account_id               string // account specified in query
    corresponding_account_id string // optional in case of deposit/withdraw
    type                     string // enum 'DEPOSIT'|'WITHDRAW'|'INTERNAL'
    direction                string // enum 'INCOMING'|'OUTGOING'
    created_at               date
}
```

Example output:
```
GET /accounts/1836981e-7bce-4356-99a5-a001073e51fe/transfers/ HTTP/1.1
Accept: */*
Accept-Encoding: gzip, deflate
Connection: keep-alive
Host: localhost:8080
User-Agent: HTTPie/2.2.0



HTTP/1.1 200 OK
Content-Length: 279
Content-Type: application/json; charset=utf-8
Date: Mon, 21 Sep 2020 11:05:53 GMT

{
    "payload": [
        {
            "account_id": "1836981e-7bce-4356-99a5-a001073e51fe",
            "amount": "1000",
            "corresponding_account_id": null,
            "created_at": "2020-09-20T08:56:20.754286Z",
            "currency_code": "USD",
            "direction": "INCOMING",
            "id": "616f2ce5-ed3b-4888-9fad-81e66fa08c26",
            "type": "DEPOSIT"
        }
    ],
    "result": "OK"
}
```
 
### GetAllAccounts

`GET <endpoint>/accounts/`

Method returns array accounts ordered by `updated_at` if count greater than 100, limit to 100
```
entity account {
    id            string
    currency_code string
    balance       decimal
    created_at    date
    updated_at    date
}
```

Example output:
```
GET /accounts/ HTTP/1.1
User-Agent: HTTPie/2.2.0
Accept-Encoding: gzip, deflate
Accept: */*
Connection: keep-alive
Host: localhost:8080

HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Mon, 21 Sep 2020 11:03:29 GMT
Content-Length: 1048

{
  "result": "OK",
  "payload": [
    {
      "id": "1836981e-7bce-4356-99a5-a001073e51fe",
      "currency_code": "USD",
      "balance": "1000",
      "created_at": "2020-09-20T08:56:20.754286Z",
      "updated_at": "2020-09-20T08:56:20.754286Z"
    },
    ....
  ]
}
```