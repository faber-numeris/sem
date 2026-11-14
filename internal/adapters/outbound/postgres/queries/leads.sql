-- name: CreateLead :one
INSERT INTO salesforce.lead (
    external_id__c,
    last_name,
    first_name,
    salutation,
    name,
    title,
    company,
    street,
    city,
    state,
    postal_code,
    country,
    phone,
    mobile_phone,
    email,
    website,
    description,
    lead_source,
    status,
    industry,
    rating,
    annual_revenue,
    number_of_employees
)
VALUES (
    sqlc.narg('external_id__c')::TEXT,
    @last_name::TEXT,
    sqlc.narg('first_name')::TEXT,
    sqlc.narg('salutation')::TEXT,
    @name::TEXT,
    sqlc.narg('title')::TEXT,
    @company::TEXT,
    sqlc.narg('street')::TEXT,
    sqlc.narg('city')::TEXT,
    sqlc.narg('state')::TEXT,
    sqlc.narg('postal_code')::TEXT,
    sqlc.narg('country')::TEXT,
    sqlc.narg('phone')::TEXT,
    sqlc.narg('mobile_phone')::TEXT,
    sqlc.narg('email')::TEXT,
    sqlc.narg('website')::TEXT,
    sqlc.narg('description')::TEXT,
    sqlc.narg('lead_source')::TEXT,
    @status::TEXT,
    sqlc.narg('industry')::TEXT,
    sqlc.narg('rating')::TEXT,
    sqlc.narg('annual_revenue')::FLOAT8,
    sqlc.narg('number_of_employees')::INT
)
RETURNING *;

-- name: GetLeadByID :one
SELECT *
FROM salesforce.lead
WHERE id = @id::TEXT
  AND is_deleted = FALSE
LIMIT 1;

-- name: GetLeadByExternalID :one
SELECT *
FROM salesforce.lead
WHERE external_id__c = @external_id::TEXT
  AND is_deleted = FALSE
LIMIT 1;

-- name: UpdateLead :one
UPDATE salesforce.lead
SET external_id__c      = COALESCE(sqlc.narg('external_id__c')::TEXT, external_id__c),
    last_name           = COALESCE(sqlc.narg('last_name')::TEXT, last_name),
    first_name          = COALESCE(sqlc.narg('first_name')::TEXT, first_name),
    name                = COALESCE(sqlc.narg('name')::TEXT, name),
    title               = COALESCE(sqlc.narg('title')::TEXT, title),
    company             = COALESCE(sqlc.narg('company')::TEXT, company),
    phone               = COALESCE(sqlc.narg('phone')::TEXT, phone),
    email               = COALESCE(sqlc.narg('email')::TEXT, email),
    website             = COALESCE(sqlc.narg('website')::TEXT, website),
    description         = COALESCE(sqlc.narg('description')::TEXT, description),
    status              = COALESCE(sqlc.narg('status')::TEXT, status),
    industry            = COALESCE(sqlc.narg('industry')::TEXT, industry),
    rating              = COALESCE(sqlc.narg('rating')::TEXT, rating),
    last_modified_date  = CURRENT_TIMESTAMP,
    system_modstamp     = CURRENT_TIMESTAMP
WHERE id = @id::TEXT
  AND is_deleted = FALSE
RETURNING *;

-- name: SoftDeleteLead :execrows
UPDATE salesforce.lead
SET is_deleted         = TRUE,
    last_modified_date = CURRENT_TIMESTAMP,
    system_modstamp    = CURRENT_TIMESTAMP
WHERE id = @id::TEXT
  AND is_deleted = FALSE;

-- name: ListLeads :many
SELECT *
FROM salesforce.lead
WHERE is_deleted = FALSE
ORDER BY created_date ASC;

-- name: ListLeadsAll :many
SELECT *
FROM salesforce.lead
ORDER BY created_date ASC;

-- name: SearchLeads :many
SELECT *
FROM salesforce.lead
WHERE is_deleted = FALSE
  AND (
    name ILIKE '%' || @term::TEXT || '%'
    OR company ILIKE '%' || @term::TEXT || '%'
    OR COALESCE(email, '') ILIKE '%' || @term::TEXT || '%'
    OR COALESCE(phone, '') ILIKE '%' || @term::TEXT || '%'
  )
ORDER BY created_date ASC;
