# API methods

## Common response format

```
Entity {
  ResultCode string // enum 'OK'|'ERROR'
  Error      string // empty if ResultCode is 'OK'
  Payload    object|array // 
}
```

## CreateInnerTransfer

Create internal transfer between the two wallet accounts
```
entity InnerTransferOrder {
	ID                string // acts as idempotency key
	SenderAccountID   string
	RecieverAccountID string
	Amount            decimal
	CurrencyCode      string 
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
entity Transfers {
    ID                     string
    CurrencyCode           string
    Amount                 decimal
    AccountID              string // account specified in query
    CorrespondingAccountID string // optional in case of deposit/withdraw
    Type                   string // enum 'DEPOSIT'|'WITHDRAW'|'INTERNAL'
    Direction              string // enum 'INCOMING'|'OUTGOING'
    CreatedAt              date
}
```
 - ErrAccountNotFound
 - ErrUnknown - 500
 
### GetAllAccounts

Method returns accounts ordered by `UpdatedAt` if count greater than 100, limit to 100
```
entity Account {
    ID           string
    CurrencyCode string
    Balance      decimal
    CreatedAt    date
    UpdatedAt    date
}
```
 - UnknownError - 500