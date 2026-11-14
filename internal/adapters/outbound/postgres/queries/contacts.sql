-- name: CreateContact :one
INSERT INTO salesforce.contact (
    external_id__c,
    account_id,
    last_name,
    first_name,
    salutation,
    name,
    mailing_street,
    mailing_city,
    mailing_state,
    mailing_postal_code,
    mailing_country,
    phone,
    fax,
    mobile_phone,
    email,
    title,
    department,
    lead_source,
    description
)
VALUES (
    sqlc.narg('external_id__c')::TEXT,
    sqlc.narg('account_id')::TEXT,
    @last_name::TEXT,
    sqlc.narg('first_name')::TEXT,
    sqlc.narg('salutation')::TEXT,
    @name::TEXT,
    sqlc.narg('mailing_street')::TEXT,
    sqlc.narg('mailing_city')::TEXT,
    sqlc.narg('mailing_state')::TEXT,
    sqlc.narg('mailing_postal_code')::TEXT,
    sqlc.narg('mailing_country')::TEXT,
    sqlc.narg('phone')::TEXT,
    sqlc.narg('fax')::TEXT,
    sqlc.narg('mobile_phone')::TEXT,
    sqlc.narg('email')::TEXT,
    sqlc.narg('title')::TEXT,
    sqlc.narg('department')::TEXT,
    sqlc.narg('lead_source')::TEXT,
    sqlc.narg('description')::TEXT
)
RETURNING *;

-- name: GetContactByID :one
SELECT *
FROM salesforce.contact
WHERE id = @id::TEXT
  AND is_deleted = FALSE
LIMIT 1;

-- name: GetContactByExternalID :one
SELECT *
FROM salesforce.contact
WHERE external_id__c = @external_id::TEXT
  AND is_deleted = FALSE
LIMIT 1;

-- name: UpdateContact :one
UPDATE salesforce.contact
SET external_id__c      = COALESCE(sqlc.narg('external_id__c')::TEXT, external_id__c),
    account_id          = COALESCE(sqlc.narg('account_id')::TEXT, account_id),
    last_name           = COALESCE(sqlc.narg('last_name')::TEXT, last_name),
    first_name          = COALESCE(sqlc.narg('first_name')::TEXT, first_name),
    name                = COALESCE(sqlc.narg('name')::TEXT, name),
    mailing_street      = COALESCE(sqlc.narg('mailing_street')::TEXT, mailing_street),
    mailing_city        = COALESCE(sqlc.narg('mailing_city')::TEXT, mailing_city),
    mailing_state       = COALESCE(sqlc.narg('mailing_state')::TEXT, mailing_state),
    mailing_postal_code = COALESCE(sqlc.narg('mailing_postal_code')::TEXT, mailing_postal_code),
    mailing_country     = COALESCE(sqlc.narg('mailing_country')::TEXT, mailing_country),
    phone               = COALESCE(sqlc.narg('phone')::TEXT, phone),
    mobile_phone        = COALESCE(sqlc.narg('mobile_phone')::TEXT, mobile_phone),
    email               = COALESCE(sqlc.narg('email')::TEXT, email),
    title               = COALESCE(sqlc.narg('title')::TEXT, title),
    department          = COALESCE(sqlc.narg('department')::TEXT, department),
    description         = COALESCE(sqlc.narg('description')::TEXT, description),
    last_modified_date  = CURRENT_TIMESTAMP,
    system_modstamp     = CURRENT_TIMESTAMP
WHERE id = @id::TEXT
  AND is_deleted = FALSE
RETURNING *;

-- name: SoftDeleteContact :execrows
UPDATE salesforce.contact
SET is_deleted         = TRUE,
    last_modified_date = CURRENT_TIMESTAMP,
    system_modstamp    = CURRENT_TIMESTAMP
WHERE id = @id::TEXT
  AND is_deleted = FALSE;

-- name: ListContacts :many
SELECT *
FROM salesforce.contact
WHERE is_deleted = FALSE
ORDER BY created_date ASC;

-- name: ListContactsAll :many
SELECT *
FROM salesforce.contact
ORDER BY created_date ASC;

-- name: SearchContacts :many
SELECT *
FROM salesforce.contact
WHERE is_deleted = FALSE
  AND (
    name ILIKE '%' || @term::TEXT || '%'
    OR COALESCE(email, '') ILIKE '%' || @term::TEXT || '%'
    OR COALESCE(phone, '') ILIKE '%' || @term::TEXT || '%'
    OR COALESCE(title, '') ILIKE '%' || @term::TEXT || '%'
    OR COALESCE(department, '') ILIKE '%' || @term::TEXT || '%'
  )
ORDER BY created_date ASC;
