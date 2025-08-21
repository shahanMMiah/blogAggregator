-- name: ResetFeeds :exec
DELETE FROM feeds;

-- name: CreateFeed :one
INSERT INTO feeds(name, url, user_id)
VALUES(
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetFeed :one
SELECT * FROM feeds WHERE url = $1 LIMIT 1;

-- name: GetFeedFromName :one
SELECT * FROM feeds WHERE name = $1 LIMIT 1;

-- name: GetFeeds :many
SELECT * FROM feeds;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET last_fetched_at = $1
WHERE url = $2;

-- name: GetNextFetchedFeed :one
SELECT * FROM feeds
INNER JOIN feed_follows ON feed_follows.user_id = $1 and feed_follows.feed_id = url
ORDER BY last_fetched_at ASC NULLS FIRST
LIMIT 1;

-- name: RemoveFeeds :exec
DELETE FROM feeds WHERE name = $1;