# wallet-api

Service stores account balances and provide one with functionality 
of transferring money from one account to another.

## Domain

## Assumptions

- All IDs here are uuid4.
- Term transfer suits better than payment
- No functionality for deposit/withdraw - just inner transfer
- Each transfer consist of one(deposit/withdraw) or two(inner) parts that
are denormalized for better read performance.

## Conventions

- 'USD' and 'usd' means the same, but we use first for inner representation
- for all enums uppercase used
- uuid4 is presented in lowercase

## Questions and features to be considered for future

- Separation for booking and spending for 2-phase transactions
- Pagination for 'all' methods
- Currencies exchange
- Deposit/Withdraw from outside world 
- Additional API methods
- Clearing idempotency keys that are older that 24h e.g.

## Things used for development

### Code

- go1.15.2
- docker and docker-compose
- PostgreSQL
- github.com/shopspring/decimal for decimal numbers processing

### Tools
- github.com/rubenv/sql-migrate - DB migrations
- github.com/golangci/golangci-lint - go linter
- github.com/cortesi/modd - watcher for restarting code after changes
- github.com/go-task/task - task runner better `make` alternative
- github.com/go-kit/kit - microservices instrumentation
- githooks - code quality management
- travis ci - CI

