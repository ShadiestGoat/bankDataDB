-- name: GetTransCount :one
SELECT COUNT(*) FROM transactions WHERE author_id = $1;

-- name: DoesTransactionExist :one
SELECT EXISTS(
    SELECT 1 FROM transactions
    WHERE
        author_id = $1
            AND
        authed_at = $2 AND settled_at = $3
            AND
        description = $4 AND amount = $5
);

-- name: GetUserUpdatedAt :one
SELECT updated_at FROM users WHERE id = $1;

-- name: DoesCategoryExist :one
SELECT EXISTS(
    SELECT 1 FROM categories
    WHERE
        author_id = $1
            AND
        id = $2
);

-- name: ResetCategoryData :exec
UPDATE categories SET name = $2, color = $3, icon = $4 WHERE id = $1;