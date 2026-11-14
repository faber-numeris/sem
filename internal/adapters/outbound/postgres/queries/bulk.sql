-- name: CreateBulkIngestJob :one
INSERT INTO salesforce.bulk_ingest_job (
    operation,
    object_name,
    external_id_field_name,
    content_type,
    column_delimiter,
    line_ending,
    api_version
)
VALUES (
    @operation::TEXT,
    @object_name::TEXT,
    sqlc.narg('external_id_field_name')::TEXT,
    @content_type::TEXT,
    @column_delimiter::TEXT,
    @line_ending::TEXT,
    @api_version::TEXT
)
RETURNING *;

-- name: GetBulkIngestJob :one
SELECT *
FROM salesforce.bulk_ingest_job
WHERE id = @id::TEXT
LIMIT 1;

-- name: UpdateBulkIngestJobState :one
UPDATE salesforce.bulk_ingest_job
SET state           = @state::TEXT,
    system_modstamp = CURRENT_TIMESTAMP
WHERE id = @id::TEXT
RETURNING *;

-- name: UpdateBulkIngestJobStats :one
UPDATE salesforce.bulk_ingest_job
SET state                    = @state::TEXT,
    number_records_processed = @number_records_processed::INT,
    number_records_failed    = @number_records_failed::INT,
    system_modstamp          = CURRENT_TIMESTAMP
WHERE id = @id::TEXT
RETURNING *;

-- name: ListBulkIngestJobs :many
SELECT *
FROM salesforce.bulk_ingest_job
ORDER BY created_date DESC;

-- name: CreateBulkIngestBatch :one
INSERT INTO salesforce.bulk_ingest_batch (
    job_id,
    csv_content
)
VALUES (
    @job_id::TEXT,
    @csv_content::TEXT
)
RETURNING *;

-- name: ListBulkIngestBatchesByJob :many
SELECT *
FROM salesforce.bulk_ingest_batch
WHERE job_id = @job_id::TEXT
ORDER BY created_date ASC;

-- name: CreateBulkIngestResult :one
INSERT INTO salesforce.bulk_ingest_result (
    job_id,
    sf_id,
    sf_created,
    sf_error,
    is_success,
    row_csv
)
VALUES (
    @job_id::TEXT,
    @sf_id::TEXT,
    @sf_created::BOOL,
    @sf_error::TEXT,
    @is_success::BOOL,
    @row_csv::TEXT
)
RETURNING *;

-- name: ListBulkIngestResults :many
SELECT *
FROM salesforce.bulk_ingest_result
WHERE job_id = @job_id::TEXT
  AND is_success = @is_success::BOOL
ORDER BY id ASC;

-- name: CreateBulkQueryJob :one
INSERT INTO salesforce.bulk_query_job (
    operation,
    query_text,
    object_name,
    content_type,
    column_delimiter,
    line_ending,
    state,
    api_version
)
VALUES (
    @operation::TEXT,
    @query_text::TEXT,
    @object_name::TEXT,
    @content_type::TEXT,
    @column_delimiter::TEXT,
    @line_ending::TEXT,
    @state::TEXT,
    @api_version::TEXT
)
RETURNING *;

-- name: GetBulkQueryJob :one
SELECT *
FROM salesforce.bulk_query_job
WHERE id = @id::TEXT
LIMIT 1;
