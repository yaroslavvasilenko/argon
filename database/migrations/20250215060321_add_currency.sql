-- +goose Up
-- +goose StatementBegin
CREATE TABLE currency_exchange_rates (
    symbol VARCHAR(10) PRIMARY KEY,          -- Валютная пара как первичный ключ
    exchange_rate DECIMAL(18,8) NOT NULL,     -- Текущий курс
    expires_at TIMESTAMP,                     -- Срок актуальности (TTL)
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE currency_exchange_rates;
-- +goose StatementEnd