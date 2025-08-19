-- +goose Up
CREATE TABLE posts(
    id UUID PRIMARY KEY NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    title TEXT NOT NULL,
    url TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL,
    published_at TIMESTAMP NOT NULL,
    feed_id TEXT NOT NULL
);

-- +goose down
DROP TABLE posts;
