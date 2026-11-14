-- name: CreateQueryCursor :one
INSERT INTO salesforce.query_cursor (
    id,
    sobject_name,
    soql_query,
    include_deleted,
    offset_value,
    batch_size,
    total_size
)
VALUES (
    @id::TEXT,
    @sobject_name::TEXT,
    @soql_query::TEXT,
    @include_deleted::BOOL,
    @offset_value::INT,
    @batch_size::INT,
    @total_size::INT
)
RETURNING *;

-- name: GetQueryCursor :one
SELECT *
FROM salesforce.query_cursor
WHERE id = @id::TEXT
LIMIT 1;

-- name: UpdateQueryCursorOffset :one
UPDATE salesforce.query_cursor
SET offset_value = @offset_value::INT
WHERE id = @id::TEXT
RETURNING *;

-- name: DeleteQueryCursor :exec
DELETE FROM salesforce.query_cursor
WHERE id = @id::TEXT;
