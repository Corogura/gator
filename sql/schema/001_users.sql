-- +goose Up
CREATE TABLE users(
    id TEXT PRIMARY KEY,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    name TEXT UNIQUE
);

-- +goose Down
DROP TABLE users;