-- +goose Up
CREATE TABLE feeds (
    id uuid PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    name VARCHAR  NOT NULL,
    url VARCHAR NOT NULL,
    user_id uuid ,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE feeds;