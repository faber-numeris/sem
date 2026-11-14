-- name: CreateCase :one
INSERT INTO salesforce.case (
    external_id__c,
    additionalinfo__c,
    sub_type__c,
    type__c,
    type_name,
    origin,
    record_type_id,
    parent_id,
    population_impacted__c,
    closed_date,
    weblinks__c,
    root_cause__c,
    impact__c,
    vip__c,
    priority,
    account_id,
    contact_id,
    recurrence__c,
    human_factor__c,
    technical_system__c,
    contact_conclusion_level_2__c,
    contact_conclusion_level_1__c,
    contact_action_level_3__c,
    contact_action_level_2__c,
    contact_action_level_1__c,
    contact_topic_level_3__c,
    contact_topic_level_2__c,
    contact_topic_level_1__c,
    components__c,
    steps_to_reproduce__c,
    expected_result__c,
    actual_result__c,
    description,
    owner_id,
    sub_status_level_3__c,
    sub_status_level_2__c,
    sub_status__c,
    status,
    created_by_id,
    subject
)
VALUES (
    sqlc.narg('external_id__c')::TEXT,
    sqlc.narg('additionalInfo__c')::TEXT,
    sqlc.narg('sub_type__c')::TEXT,
    sqlc.narg('type__c')::TEXT,
    sqlc.narg('type_name')::TEXT,
    sqlc.narg('origin')::TEXT,
    sqlc.narg('record_type_id')::TEXT,
    sqlc.narg('parent_id')::TEXT,
    sqlc.narg('population_impacted__c')::TEXT,
    sqlc.narg('closed_date')::TIMESTAMPTZ,
    sqlc.narg('weblinks__c')::TEXT,
    sqlc.narg('root_cause__c')::TEXT,
    sqlc.narg('impact__c')::TEXT,
    sqlc.narg('vip__c')::BOOL,
    @priority::TEXT,
    sqlc.narg('account_id')::TEXT,
    sqlc.narg('contact_id')::TEXT,
    sqlc.narg('recurrence__c')::TEXT,
    sqlc.narg('human_factor__c')::TEXT,
    sqlc.narg('technical_system__c')::TEXT,
    sqlc.narg('contact_conclusion_level_2__c')::TEXT,
    sqlc.narg('contact_conclusion_level_1__c')::TEXT,
    sqlc.narg('contact_action_level_3__c')::TEXT,
    sqlc.narg('contact_action_level_2__c')::TEXT,
    sqlc.narg('contact_action_level_1__c')::TEXT,
    sqlc.narg('contact_topic_level_3__c')::TEXT,
    sqlc.narg('contact_topic_level_2__c')::TEXT,
    sqlc.narg('contact_topic_level_1__c')::TEXT,
    sqlc.narg('components__c')::TEXT,
    sqlc.narg('steps_to_reproduce__c')::TEXT,
    sqlc.narg('expected_result__c')::TEXT,
    sqlc.narg('actual_result__c')::TEXT,
    sqlc.narg('description')::TEXT,
    sqlc.narg('owner_id')::TEXT,
    sqlc.narg('sub_status_level_3__c')::TEXT,
    sqlc.narg('sub_status_level_2__c')::TEXT,
    sqlc.narg('sub_status__c')::TEXT,
    sqlc.narg('status')::TEXT,
    sqlc.narg('created_by_id')::TEXT,
    sqlc.narg('subject')::TEXT
)
RETURNING *;

-- name: SelectCase :one
SELECT *
FROM salesforce.case
WHERE id = @id::TEXT
  AND is_deleted = FALSE
LIMIT 1;

-- name: GetCaseByExternalID :one
SELECT *
FROM salesforce.case
WHERE external_id__c = @external_id::TEXT
  AND is_deleted = FALSE
LIMIT 1;

-- name: UpdateCase :one
UPDATE salesforce.case
SET external_id__c                = COALESCE(sqlc.narg('external_id__c')::TEXT, external_id__c),
    additionalinfo__c             = COALESCE(sqlc.narg('additionalInfo__c')::TEXT, additionalinfo__c),
    sub_type__c                   = COALESCE(sqlc.narg('sub_type__c')::TEXT, sub_type__c),
    type__c                       = COALESCE(sqlc.narg('type__c')::TEXT, type__c),
    type_name                     = COALESCE(sqlc.narg('type_name')::TEXT, type_name),
    origin                        = COALESCE(sqlc.narg('origin')::TEXT, origin),
    record_type_id                = COALESCE(sqlc.narg('record_type_id')::TEXT, record_type_id),
    parent_id                     = COALESCE(sqlc.narg('parent_id')::TEXT, parent_id),
    population_impacted__c        = COALESCE(sqlc.narg('population_impacted__c')::TEXT, population_impacted__c),
    closed_date                   = COALESCE(sqlc.narg('closed_date')::TIMESTAMPTZ, closed_date),
    weblinks__c                   = COALESCE(sqlc.narg('weblinks__c')::TEXT, weblinks__c),
    root_cause__c                 = COALESCE(sqlc.narg('root_cause__c')::TEXT, root_cause__c),
    impact__c                     = COALESCE(sqlc.narg('impact__c')::TEXT, impact__c),
    vip__c                        = COALESCE(sqlc.narg('vip__c')::BOOL, vip__c),
    priority                      = COALESCE(sqlc.narg('priority')::TEXT, priority),
    account_id                    = COALESCE(sqlc.narg('account_id')::TEXT, account_id),
    contact_id                    = COALESCE(sqlc.narg('contact_id')::TEXT, contact_id),
    description                   = COALESCE(sqlc.narg('description')::TEXT, description),
    status                        = COALESCE(sqlc.narg('status')::TEXT, status),
    subject                       = COALESCE(sqlc.narg('subject')::TEXT, subject),
    last_modified_date            = CURRENT_TIMESTAMP,
    system_modstamp               = CURRENT_TIMESTAMP
WHERE id = @id::TEXT
  AND is_deleted = FALSE
RETURNING *;

-- name: SoftDeleteCase :execrows
UPDATE salesforce.case
SET is_deleted         = TRUE,
    last_modified_date = CURRENT_TIMESTAMP,
    system_modstamp    = CURRENT_TIMESTAMP
WHERE id = @id::TEXT
  AND is_deleted = FALSE;

-- name: ListCases :many
SELECT *
FROM salesforce.case
WHERE is_deleted = FALSE
ORDER BY created_date ASC;

-- name: ListCasesAll :many
SELECT *
FROM salesforce.case
ORDER BY created_date ASC;

-- name: SearchCases :many
SELECT *
FROM salesforce.case
WHERE is_deleted = FALSE
  AND (
    case_number ILIKE '%' || @term::TEXT || '%'
    OR COALESCE(subject, '') ILIKE '%' || @term::TEXT || '%'
    OR COALESCE(description, '') ILIKE '%' || @term::TEXT || '%'
  )
ORDER BY created_date ASC;
