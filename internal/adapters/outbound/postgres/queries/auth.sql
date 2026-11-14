-- name: GetOAuthClientByClientID :one
SELECT *
FROM salesforce.oauth_client
WHERE client_id = @client_id::TEXT
LIMIT 1;

-- name: GetUserByUsername :one
SELECT *
FROM salesforce.salesforce_user
WHERE username = @username::TEXT
LIMIT 1;

-- name: GetUserByID :one
SELECT *
FROM salesforce.salesforce_user
WHERE id = @id::TEXT
LIMIT 1;

-- name: GetDefaultUser :one
SELECT *
FROM salesforce.salesforce_user
WHERE is_active = TRUE
ORDER BY created_date ASC
LIMIT 1;

-- name: CreateAuthCode :one
INSERT INTO salesforce.oauth_auth_code (
    code,
    client_id,
    user_id,
    redirect_uri,
    scope,
    expires_at
)
VALUES (
    @code::TEXT,
    @client_id::TEXT,
    @user_id::TEXT,
    @redirect_uri::TEXT,
    @scope::TEXT,
    @expires_at::TIMESTAMPTZ
)
RETURNING *;

-- name: GetAuthCode :one
SELECT *
FROM salesforce.oauth_auth_code
WHERE code = @code::TEXT
LIMIT 1;

-- name: MarkAuthCodeUsed :exec
UPDATE salesforce.oauth_auth_code
SET is_used = TRUE
WHERE code = @code::TEXT;

-- name: CreateOAuthToken :one
INSERT INTO salesforce.oauth_token (
    access_token,
    refresh_token,
    token_type,
    client_id,
    user_id,
    scope,
    instance_url,
    id_url,
    signature,
    issued_at,
    expires_at
)
VALUES (
    @access_token::TEXT,
    sqlc.narg('refresh_token')::TEXT,
    @token_type::TEXT,
    @client_id::TEXT,
    @user_id::TEXT,
    @scope::TEXT,
    @instance_url::TEXT,
    @id_url::TEXT,
    @signature::TEXT,
    @issued_at::BIGINT,
    @expires_at::TIMESTAMPTZ
)
RETURNING *;

-- name: GetOAuthTokenByAccessToken :one
SELECT *
FROM salesforce.oauth_token
WHERE access_token = @access_token::TEXT
LIMIT 1;

-- name: GetOAuthTokenByRefreshToken :one
SELECT *
FROM salesforce.oauth_token
WHERE refresh_token = @refresh_token::TEXT
LIMIT 1;

-- name: GetOAuthTokenByAnyToken :one
SELECT *
FROM salesforce.oauth_token
WHERE access_token = @token::TEXT
   OR refresh_token = @token::TEXT
LIMIT 1;

-- name: RevokeOAuthToken :exec
UPDATE salesforce.oauth_token
SET is_revoked = TRUE
WHERE access_token = @token::TEXT
   OR refresh_token = @token::TEXT;
