-- name: ListOrgLimits :many
SELECT *
FROM salesforce.org_limit
ORDER BY limit_name ASC;

-- name: ListSObjectDefinitions :many
SELECT *
FROM salesforce.sobject_definition
WHERE is_tooling = @is_tooling::BOOL
ORDER BY name ASC;

-- name: GetSObjectDefinition :one
SELECT *
FROM salesforce.sobject_definition
WHERE lower(name) = lower(@name::TEXT)
LIMIT 1;

-- name: ListSObjectFieldDefinitions :many
SELECT *
FROM salesforce.sobject_field_definition
WHERE lower(sobject_name) = lower(@sobject_name::TEXT)
ORDER BY id ASC;
