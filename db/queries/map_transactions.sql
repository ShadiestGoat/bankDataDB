-- name: TransMapsRmNames :exec
WITH deleted AS (
    DELETE FROM mapped_transactions
    WHERE mapping_id = $1
    RETURNING trans_id
) UPDATE transactions
SET resolved_name = NULL
FROM deleted
WHERE id = deleted.trans_id;

-- name: TransMapsRmCategories :exec
WITH deleted AS (
    DELETE FROM mapped_transactions
    WHERE mapping_id = $1
    RETURNING trans_id
) UPDATE transactions
SET resolved_category = NULL
FROM deleted
WHERE id = deleted.trans_id;