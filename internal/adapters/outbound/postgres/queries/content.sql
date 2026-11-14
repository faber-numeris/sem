-- name: CreateContentDocument :one
INSERT INTO salesforce.content_document (type, title)
VALUES (sqlc.narg('type')::TEXT, sqlc.narg('title')::TEXT)
RETURNING *;

-- name: GetContentDocument :one
SELECT *
FROM salesforce.content_document
WHERE id = @id::TEXT
  AND is_deleted = FALSE
LIMIT 1;

-- name: UpdateContentDocument :one
UPDATE salesforce.content_document
SET type  = COALESCE(sqlc.narg('type')::TEXT, type),
    title = COALESCE(sqlc.narg('title')::TEXT, title)
WHERE id = @id::TEXT
  AND is_deleted = FALSE
RETURNING *;

-- name: SoftDeleteContentDocument :execrows
UPDATE salesforce.content_document
SET is_deleted = TRUE
WHERE id = @id::TEXT
  AND is_deleted = FALSE;

-- name: ListContentDocuments :many
SELECT *
FROM salesforce.content_document
WHERE is_deleted = FALSE
ORDER BY created_date ASC;

-- name: CreateContentVersion :one
INSERT INTO salesforce.content_version (
    path_on_client,
    title,
    version_data,
    content_document_id,
    version_data_url,
    file_type
)
VALUES (
    @path_on_client::TEXT,
    @title::TEXT,
    @version_data::TEXT,
    @content_document_id::TEXT,
    sqlc.narg('version_data_url')::TEXT,
    sqlc.narg('file_type')::TEXT
)
RETURNING *;

-- name: GetContentVersion :one
SELECT *
FROM salesforce.content_version
WHERE id = @id::TEXT
  AND COALESCE(is_deleted, FALSE) = FALSE
LIMIT 1;

-- name: SoftDeleteContentVersion :execrows
UPDATE salesforce.content_version
SET is_deleted = TRUE
WHERE id = @id::TEXT
  AND COALESCE(is_deleted, FALSE) = FALSE;

-- name: ListContentVersions :many
SELECT *
FROM salesforce.content_version
WHERE COALESCE(is_deleted, FALSE) = FALSE
ORDER BY created_date ASC;

-- name: CreateContentDocumentLink :one
INSERT INTO salesforce.content_document_link (
    content_document_id,
    linked_entity_id,
    share_type,
    visibility
)
VALUES (
    @content_document_id::TEXT,
    @linked_entity_id::TEXT,
    @share_type::TEXT,
    @visibility::TEXT
)
RETURNING *;

-- name: GetContentDocumentLink :one
SELECT *
FROM salesforce.content_document_link
WHERE id = @id::TEXT
  AND is_deleted = FALSE
LIMIT 1;

-- name: SoftDeleteContentDocumentLink :execrows
UPDATE salesforce.content_document_link
SET is_deleted = TRUE
WHERE id = @id::TEXT
  AND is_deleted = FALSE;

-- name: ListContentDocumentLinks :many
SELECT *
FROM salesforce.content_document_link
WHERE is_deleted = FALSE
ORDER BY id ASC;
