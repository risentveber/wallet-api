# wallet-api
[![Build Status](https://travis-ci.com/risentveber/wallet-api.svg?branch=master)](https://travis-ci.com/risentveber/wallet-api)

Service stores account balances and provide one with functionality 
of transferring money from one account to another.

## Domain

For detailed API documentation see `/docs/API.md`. You can find DDD like description
at code-level in `services/transfers/domain.go`.

## Business assumptions

- All IDs here are uuid4.
- Term transfer suits better than payment.
- No functionality for deposit/withdraw - just inner transfer.
- Each transfer consist of one(deposit/withdraw) or two(inner) parts that
are denormalized (for better read performance).
- Transfer amount rounding without any error if there is extra precision specified.

## Business conventions

- 'USD' and 'usd' means the same, but we use first for inner representation.
- Transfer operations positive-idempotent(means after succeeded 
once it will always return 'OK' without duplication).

## Tech assumptions

- Service simplified - there is no prometheus metrics, tracing and extra logging.
- It's supposed to run in k8s - there is no service discovery logic.
- For supported configuration flags see `./cmd/api/config.go`.
- When started server tries connect to DB until succeed or 
retries count exceeded(see configuration options).
- To stop server gracefully you need to send `SIGHUP`, and it will try do it
in time specified via configuration.

## DB layout

![DB Schema](/docs/schema-db.png?raw=true "DB schema used")

## Configuration
```
Usage of /bin/api:
  -db string
    	db connections credentials
  -dbRetryCount uint
    	retry count for connecting to db (default 10)
  -dbRetryTimeout duration
    	retry timeout for connecting to db (default 2s)
  -logLevel string
    	debug|info|warn|error (default "info")
  -port string
    	port (default "8080")
  -shutdownTimeout duration
    	graceful shutdown timeout (default 10s)
```

## Questions and features to be considered for future

- Separation for booking and spending for 2-phase transactions.
- Pagination for 'all' methods.
- Currencies exchange.
- Deposit/Withdraw from outside world.
- Additional API methods.

## How to run code or develop

### Preparations

You need install:
- `docker` and `docker-compose`
- https://taskfile.dev

### Run locally

```bash
task build_dev_tools # builds tooling image
task up # run all in docker compose
# wait some time until PostgreSQL will statred
task migrate # apply migrations inside docker-compose started by task up
task -l # for available tasks for development
```

### Development
Some task commands may require tools that presented in 
docker image that can build from `tools/Dockerfile-tools` via `task build_dev_tools`.

Attaches git hooks for running test/linter in development process lifecycle,
also prepending each commit msg with `[<branch_name>] ` prefix for later investigations 
via git blame.
```bash
task attach_hooks
```
After you start and change something modd rebuild and restart app via `./tools/modd.conf`
There is also `./integration/init.sql` file that maybe useful if you decide to populate DB with test data.

### Tests

```bash
task test # if have go installed or inside go-dev-tools docker image
task run_integration_tests # runs integration tests via docker-compose 
```

### Build Docker image for deployment 

Be aware you need to run migrations before image deploy (e.g. via k8s job)

```bash
IMAGE_NAME=wallet-api:1.0.0 task build_api_image
```

## Things used

### Code

- go1.15.2.
- docker and docker-compose.
- PostgreSQL.
- https://github.com/shopspring/decimal - decimal numbers processing.
- https://github.com/gorilla/mux - http routing.
- https://github.com/oklog/run - for managing top-level gorutines.


### Tools
- https://github.com/DATA-DOG/go-sqlmock - SQL mocks for unit tests.
- https://github.com/stretchr/testify/assert - test assertions.
- https://github.com/rubenv/sql-migrate - DB migrations.
- https://github.com/golangci/golangci-lint - go linter.
- https://github.com/cortesi/modd - watcher for restarting code after changes.
- https://github.com/go-task/task - task runner better `make` alternative.
- https://github.com/go-kit/kit - microservices instrumentation.
- githooks - local code quality management.
- travis ci - CI.

