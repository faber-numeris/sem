-- name: CreateCustomObject :one
INSERT INTO salesforce.custom_object__c (
    external_id__c,
    name,
    field_to_update__c,
    description__c,
    status__c
)
VALUES (
    sqlc.narg('external_id__c')::TEXT,
    @name::TEXT,
    sqlc.narg('field_to_update__c')::TEXT,
    sqlc.narg('description__c')::TEXT,
    sqlc.narg('status__c')::TEXT
)
RETURNING *;

-- name: GetCustomObjectByID :one
SELECT *
FROM salesforce.custom_object__c
WHERE id = @id::TEXT
  AND is_deleted = FALSE
LIMIT 1;

-- name: GetCustomObjectByExternalID :one
SELECT *
FROM salesforce.custom_object__c
WHERE external_id__c = @external_id::TEXT
  AND is_deleted = FALSE
LIMIT 1;

-- name: UpdateCustomObject :one
UPDATE salesforce.custom_object__c
SET external_id__c     = COALESCE(sqlc.narg('external_id__c')::TEXT, external_id__c),
    name               = COALESCE(sqlc.narg('name')::TEXT, name),
    field_to_update__c = COALESCE(sqlc.narg('field_to_update__c')::TEXT, field_to_update__c),
    description__c     = COALESCE(sqlc.narg('description__c')::TEXT, description__c),
    status__c          = COALESCE(sqlc.narg('status__c')::TEXT, status__c),
    last_modified_date = CURRENT_TIMESTAMP,
    system_modstamp    = CURRENT_TIMESTAMP
WHERE id = @id::TEXT
  AND is_deleted = FALSE
RETURNING *;

-- name: SoftDeleteCustomObject :execrows
UPDATE salesforce.custom_object__c
SET is_deleted         = TRUE,
    last_modified_date = CURRENT_TIMESTAMP,
    system_modstamp    = CURRENT_TIMESTAMP
WHERE id = @id::TEXT
  AND is_deleted = FALSE;

-- name: ListCustomObjects :many
SELECT *
FROM salesforce.custom_object__c
WHERE is_deleted = FALSE
ORDER BY created_date ASC;

-- name: ListCustomObjectsAll :many
SELECT *
FROM salesforce.custom_object__c
ORDER BY created_date ASC;

-- name: SearchCustomObjects :many
SELECT *
FROM salesforce.custom_object__c
WHERE is_deleted = FALSE
  AND (
    name ILIKE '%' || @term::TEXT || '%'
    OR COALESCE(field_to_update__c, '') ILIKE '%' || @term::TEXT || '%'
    OR COALESCE(description__c, '') ILIKE '%' || @term::TEXT || '%'
  )
ORDER BY created_date ASC;
