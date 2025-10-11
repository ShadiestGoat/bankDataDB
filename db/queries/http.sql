-- name: ExtGetCategories :many
SELECT id, color, icon, name FROM categories WHERE author_id = $1;

-- name: ExtDelCategory :execrows
DELETE FROM categories WHERE author_id = $1 AND id = $2;
