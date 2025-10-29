-- name: TransMapsCleanAll :exec
WITH deleted AS (
    DELETE FROM mapped_transactions
    WHERE mapping_id = $1
    RETURNING trans_id, updated_name
), names AS (
    UPDATE transactions
    SET resolved_name = NULL
    FROM deleted
    WHERE id = deleted.trans_id AND updated_name IS TRUE
) UPDATE transactions
SET resolved_category = NULL
FROM deleted
WHERE id = deleted.trans_id AND updated_name IS FALSE;

-- name: TransMapsOrphanAll :exec
DELETE FROM mapped_transactions WHERE mapping_id = $1;

-- name: TransMapsOrphanNames :exec
DELETE FROM mapped_transactions WHERE mapping_id = $1 AND updated_name IS TRUE;

-- name: TransMapsOrphanCategories :exec
DELETE FROM mapped_transactions WHERE mapping_id = $1 AND updated_name IS FALSE;

-- name: TransMapsUpdateLinkedNames :exec
UPDATE transactions AS t
SET resolved_name = $2
FROM mapped_transactions AS mp
WHERE
    mp.mapping_id = $1
        AND
    t.id = mp.trans_id
        AND
    updated_name IS TRUE;

-- name: TransMapsUpdateLinkedCategories :exec
UPDATE transactions AS t
SET resolved_category = $2
FROM mapped_transactions AS mp
WHERE
    mp.mapping_id = $1
        AND
    t.id = mp.trans_id
        AND
    updated_name IS FALSE;
