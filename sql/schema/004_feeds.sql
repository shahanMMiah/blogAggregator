-- +goose Up
CREATE TABLE feeds(
    name TEXT NOT NULL,
    url TEXT UNIQUE NOT NULL,
    user_id UUID REFERENCES users (id) ON DELETE CASCADE NOT NULL
);

-- +goose Down
DROP TABLE feeds;