-- +migrate Up
CREATE TABLE currencies
(
    code      varchar(4) PRIMARY KEY,
    precision int not null
);
INSERT INTO currencies
VALUES ('USD', 2),
       ('EUR', 2),
       ('BTC', 8);

CREATE TABLE accounts
(
    id            uuid PRIMARY KEY,
    currency_code varchar(4) not null references currencies (code),
    balance       decimal    not null default 0.0 check ( balance >= 0 ),
    created_at    date       not null default now(),
    updated_at    date       not null default now()
);

CREATE INDEX accounts_by_updated_at on accounts (updated_at);

CREATE TYPE transfer_type AS ENUM ('DEPOSIT', 'WITHDRAW', 'INTERNAL');

CREATE TABLE transfers
(
    id            uuid PRIMARY KEY,
    type          transfer_type,
    amount        decimal    not null check ( amount > 0 ),
    currency_code varchar(4) not null references currencies (code),
    created_at    date       not null default now()
);

CREATE TYPE direction_type AS ENUM ('INCOMING', 'OUTGOING');

CREATE TABLE transfer_parts
(
    transfer_id              uuid           not null references transfers (id),
    account_id               uuid           not null references accounts (id),
    corresponding_account_id uuid references accounts (id), -- may be null in case of Deposit/Withdrawal
    direction                direction_type not null,
    PRIMARY KEY (transfer_id, account_id)
);
CREATE INDEX transfer_parts_by_account_id on transfer_parts (account_id);
CREATE INDEX transfer_parts_by_transfer_id on transfer_parts (transfer_id);

-- +migrate Down
DROP INDEX transfer_parts_by_transfer_id;
DROP INDEX transfer_parts_by_account_id;
DROP TABLE transfer_parts;

DROP TABLE transfers;
DROP TYPE transfer_type;
DROP INDEX accounts_by_updated_at;

DROP TABLE accounts;
DROP TABLE currencies;