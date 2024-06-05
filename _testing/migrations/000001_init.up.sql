CREATE SCHEMA test;

CREATE TABLE test.account
(
    account_id         SERIAL PRIMARY KEY,
    account_name       VARCHAR     NOT NULL,
    account_created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO test.account (account_name)
VALUES ('test a'),
       ('test b'),
       ('test c');
