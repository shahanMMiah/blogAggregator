-- name: CreateFeedFollow :one

WITH feed_folow_insert AS (
    INSERT INTO feed_follows(id, created_at, updated_at, user_id, feed_id)
    VALUES(
        $1,
        $2,
        $3,
        $4,
        $5
    )
    RETURNING *
) 
SELECT feed_folow_insert.*,
    feeds.name AS feed_name,
    users.name AS user_name
    FROM feed_folow_insert
    INNER JOIN feeds ON feed_folow_insert.feed_id = feeds.url
    INNER JOIN users ON feed_folow_insert.user_id = users.id;