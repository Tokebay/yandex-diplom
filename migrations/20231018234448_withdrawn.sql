-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS withdrawals (
    id SERIAL PRIMARY KEY,
    order_id VARCHAR(255) NOT NULL UNIQUE,
    user_id integer REFERENCES users(id),
    accrual decimal,
    created_at TIMESTAMPTZ NOT NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS withdrawals;
-- +goose StatementEnd
