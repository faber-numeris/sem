-- name: ListApexClasses :many
SELECT *
FROM salesforce.apex_class
ORDER BY name ASC;

-- name: GetApexClassByID :one
SELECT *
FROM salesforce.apex_class
WHERE id = @id::TEXT
LIMIT 1;

-- name: CreateApexClass :one
INSERT INTO salesforce.apex_class (
    name,
    body,
    api_version,
    status,
    is_valid,
    length_without_comments
)
VALUES (
    @name::TEXT,
    @body::TEXT,
    @api_version::FLOAT8,
    @status::TEXT,
    @is_valid::BOOL,
    @length_without_comments::INT
)
RETURNING *;

-- name: ListApexTriggers :many
SELECT *
FROM salesforce.apex_trigger
ORDER BY name ASC;

-- name: GetApexTriggerByID :one
SELECT *
FROM salesforce.apex_trigger
WHERE id = @id::TEXT
LIMIT 1;

-- name: ListCustomFields :many
SELECT *
FROM salesforce.custom_field
ORDER BY developer_name ASC;

-- name: ListValidationRules :many
SELECT *
FROM salesforce.validation_rule
ORDER BY validation_name ASC;

-- name: CreateApexTestRun :one
INSERT INTO salesforce.apex_test_run (
    class_names,
    test_level,
    status,
    completed_date
)
VALUES (
    @class_names::TEXT,
    @test_level::TEXT,
    @status::TEXT,
    CURRENT_TIMESTAMP
)
RETURNING *;
