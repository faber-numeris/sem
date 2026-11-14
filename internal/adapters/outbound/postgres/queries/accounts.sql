-- name: CreateAccount :one
INSERT INTO salesforce.account (
    external_id__c,
    name,
    type,
    parent_id,
    billing_street,
    billing_city,
    billing_state,
    billing_postal_code,
    billing_country,
    shipping_street,
    shipping_city,
    shipping_state,
    shipping_postal_code,
    shipping_country,
    phone,
    fax,
    account_number,
    website,
    industry,
    annual_revenue,
    number_of_employees,
    ownership,
    ticker_symbol,
    description,
    rating,
    site
)
VALUES (
    sqlc.narg('external_id__c')::TEXT,
    @name::TEXT,
    sqlc.narg('type')::TEXT,
    sqlc.narg('parent_id')::TEXT,
    sqlc.narg('billing_street')::TEXT,
    sqlc.narg('billing_city')::TEXT,
    sqlc.narg('billing_state')::TEXT,
    sqlc.narg('billing_postal_code')::TEXT,
    sqlc.narg('billing_country')::TEXT,
    sqlc.narg('shipping_street')::TEXT,
    sqlc.narg('shipping_city')::TEXT,
    sqlc.narg('shipping_state')::TEXT,
    sqlc.narg('shipping_postal_code')::TEXT,
    sqlc.narg('shipping_country')::TEXT,
    sqlc.narg('phone')::TEXT,
    sqlc.narg('fax')::TEXT,
    sqlc.narg('account_number')::TEXT,
    sqlc.narg('website')::TEXT,
    sqlc.narg('industry')::TEXT,
    sqlc.narg('annual_revenue')::FLOAT8,
    sqlc.narg('number_of_employees')::INT,
    sqlc.narg('ownership')::TEXT,
    sqlc.narg('ticker_symbol')::TEXT,
    sqlc.narg('description')::TEXT,
    sqlc.narg('rating')::TEXT,
    sqlc.narg('site')::TEXT
)
RETURNING *;

-- name: GetAccountByID :one
SELECT *
FROM salesforce.account
WHERE id = @id::TEXT
  AND is_deleted = FALSE
LIMIT 1;

-- name: GetAccountByExternalID :one
SELECT *
FROM salesforce.account
WHERE external_id__c = @external_id::TEXT
  AND is_deleted = FALSE
LIMIT 1;

-- name: UpdateAccount :one
UPDATE salesforce.account
SET external_id__c      = COALESCE(sqlc.narg('external_id__c')::TEXT, external_id__c),
    name                = COALESCE(sqlc.narg('name')::TEXT, name),
    type                = COALESCE(sqlc.narg('type')::TEXT, type),
    parent_id           = COALESCE(sqlc.narg('parent_id')::TEXT, parent_id),
    billing_street      = COALESCE(sqlc.narg('billing_street')::TEXT, billing_street),
    billing_city        = COALESCE(sqlc.narg('billing_city')::TEXT, billing_city),
    billing_state       = COALESCE(sqlc.narg('billing_state')::TEXT, billing_state),
    billing_postal_code = COALESCE(sqlc.narg('billing_postal_code')::TEXT, billing_postal_code),
    billing_country     = COALESCE(sqlc.narg('billing_country')::TEXT, billing_country),
    phone               = COALESCE(sqlc.narg('phone')::TEXT, phone),
    website             = COALESCE(sqlc.narg('website')::TEXT, website),
    industry            = COALESCE(sqlc.narg('industry')::TEXT, industry),
    annual_revenue      = COALESCE(sqlc.narg('annual_revenue')::FLOAT8, annual_revenue),
    number_of_employees = COALESCE(sqlc.narg('number_of_employees')::INT, number_of_employees),
    description         = COALESCE(sqlc.narg('description')::TEXT, description),
    rating              = COALESCE(sqlc.narg('rating')::TEXT, rating),
    site                = COALESCE(sqlc.narg('site')::TEXT, site),
    last_modified_date  = CURRENT_TIMESTAMP,
    system_modstamp     = CURRENT_TIMESTAMP
WHERE id = @id::TEXT
  AND is_deleted = FALSE
RETURNING *;

-- name: SoftDeleteAccount :execrows
UPDATE salesforce.account
SET is_deleted         = TRUE,
    last_modified_date = CURRENT_TIMESTAMP,
    system_modstamp    = CURRENT_TIMESTAMP
WHERE id = @id::TEXT
  AND is_deleted = FALSE;

-- name: ListAccounts :many
SELECT *
FROM salesforce.account
WHERE is_deleted = FALSE
ORDER BY created_date ASC;

-- name: ListAccountsAll :many
SELECT *
FROM salesforce.account
ORDER BY created_date ASC;

-- name: SearchAccounts :many
SELECT *
FROM salesforce.account
WHERE is_deleted = FALSE
  AND (
    name ILIKE '%' || @term::TEXT || '%'
    OR COALESCE(phone, '') ILIKE '%' || @term::TEXT || '%'
    OR COALESCE(website, '') ILIKE '%' || @term::TEXT || '%'
    OR COALESCE(industry, '') ILIKE '%' || @term::TEXT || '%'
    OR COALESCE(description, '') ILIKE '%' || @term::TEXT || '%'
  )
ORDER BY created_date ASC;
