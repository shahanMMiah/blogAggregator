-- +goose Up
CREATE TABLE feed_follows(
    id UUID PRIMARY KEY NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id UUID REFERENCES users (id) ON DELETE CASCADE NOT NULL,
    feed_id TEXT REFERENCES feeds(url) ON DELETE CASCADE NOT NULL,
    UNIQUE(user_id, feed_id)
);

-- +goose Down
DROP TABLE feed_follows;