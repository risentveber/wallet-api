# API methods

## Common response format

All json responses are with `snake_case` field names format

```
entity common_response {
  result_code string // enum 'OK'|'ERROR'
  error       string // empty if ResultCode is 'OK'
  payload     object|array // 
}
```

## CreateInnerTransfer

Create internal transfer between the two wallet accounts
```
entity inner_transfer_order {
	id                  string // acts as idempotency key
	sender_account_id   string
	reciever_account_id string
	amount              decimal
	currency_code       string 
}
```
- `from==to` - ErrAccountsMustBeDifferent
- `wrong precision` - Auto top-rounding to precision
- `amount>0` - ErrAmountMustBePositive
- `from.amount<amount` - ErrInsufficientFunds
- `currency not in supported set` - ErrUnsupportedCurrency
- `no from account` - ErrSenderNotExists
- `no to account` - ErrReceiverNotExists
- `sender account has different currency` - ErrSenderWrongCurrency
- `sender account has different currency` - ErrReceiverWrongCurrency
- `idempotency key duplication` - if previous operation succeeded - sends ok
 - UnknownError - 500 k
 
### GetPaymentsByAccountID

Method returns transfers ordered by `UpdatedAt` if the count greater than 100, limit to it 100
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
 - ErrAccountNotFound
 - ErrUnknown - 500
 
### GetAllAccounts

Method returns accounts ordered by `updated_at` if count greater than 100, limit to 100
```
entity account {
    id            string
    currency_code string
    balance       decimal
    created_at    date
    updated_at    date
}
```
 - UnknownError - 500