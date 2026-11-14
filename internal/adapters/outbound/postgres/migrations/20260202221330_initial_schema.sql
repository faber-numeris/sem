-- +goose Up
CREATE SCHEMA IF NOT EXISTS salesforce;

-- =============================================================================
-- 1. OAUTH 2.0 & AUTHENTICATION TABLES
-- =============================================================================
CREATE TABLE salesforce.salesforce_user
(
    id                 VARCHAR(18) PRIMARY KEY DEFAULT generate_sf_id('005'),
    organization_id    VARCHAR(18) NOT NULL    DEFAULT '00D000000000001EAA',
    username           TEXT UNIQUE NOT NULL,
    email              TEXT UNIQUE NOT NULL,
    password_hash      TEXT        NOT NULL,
    name               TEXT        NOT NULL,
    first_name         TEXT        NOT NULL    DEFAULT '',
    last_name          TEXT        NOT NULL    DEFAULT '',
    nickname           TEXT        NOT NULL    DEFAULT '',
    picture_url        TEXT        NOT NULL    DEFAULT '',
    profile_id         VARCHAR(18) NOT NULL    DEFAULT '00e000000000001AAA',
    user_type          TEXT        NOT NULL    DEFAULT 'Standard',
    language           TEXT        NOT NULL    DEFAULT 'en_US',
    locale             TEXT        NOT NULL    DEFAULT 'en_US',
    utc_offset         INT         NOT NULL    DEFAULT -18000000,
    is_active          BOOLEAN     NOT NULL    DEFAULT TRUE,
    created_date       TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    last_modified_date TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE salesforce.oauth_client
(
    id            VARCHAR(18) PRIMARY KEY DEFAULT generate_sf_id('0H4'),
    client_id     TEXT UNIQUE NOT NULL,
    client_secret TEXT        NOT NULL,
    name          TEXT        NOT NULL,
    redirect_uri  TEXT        NOT NULL,
    scopes        TEXT        NOT NULL    DEFAULT 'api refresh_token openid',
    is_active     BOOLEAN     NOT NULL    DEFAULT TRUE,
    created_date  TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE salesforce.oauth_auth_code
(
    code         TEXT PRIMARY KEY     DEFAULT generate_ulid(),
    client_id    TEXT        NOT NULL REFERENCES salesforce.oauth_client (client_id) ON DELETE CASCADE,
    user_id      VARCHAR(18) NOT NULL REFERENCES salesforce.salesforce_user (id) ON DELETE CASCADE,
    redirect_uri TEXT        NOT NULL,
    scope        TEXT        NOT NULL DEFAULT 'api',
    is_used      BOOLEAN     NOT NULL DEFAULT FALSE,
    expires_at   TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE salesforce.oauth_token
(
    id            VARCHAR(26) PRIMARY KEY DEFAULT generate_ulid(),
    access_token  TEXT UNIQUE NOT NULL,
    refresh_token TEXT,
    token_type    TEXT        NOT NULL    DEFAULT 'Bearer',
    client_id     TEXT        NOT NULL REFERENCES salesforce.oauth_client (client_id) ON DELETE CASCADE,
    user_id       VARCHAR(18) NOT NULL REFERENCES salesforce.salesforce_user (id) ON DELETE CASCADE,
    scope         TEXT        NOT NULL    DEFAULT 'api',
    instance_url  TEXT        NOT NULL,
    id_url        TEXT        NOT NULL,
    signature     TEXT        NOT NULL    DEFAULT '',
    issued_at     BIGINT      NOT NULL,
    expires_at    TIMESTAMPTZ NOT NULL,
    is_revoked    BOOLEAN     NOT NULL    DEFAULT FALSE,
    created_at    TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- 2. METADATA & ORG LIMITS TABLES
-- =============================================================================
CREATE TABLE salesforce.org_limit
(
    limit_name      TEXT PRIMARY KEY,
    max_limit       INT         NOT NULL,
    remaining_limit INT         NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE salesforce.sobject_definition
(
    name          TEXT PRIMARY KEY,
    label         TEXT        NOT NULL,
    label_plural  TEXT        NOT NULL,
    key_prefix    VARCHAR(3)  NOT NULL,
    custom        BOOLEAN     NOT NULL DEFAULT FALSE,
    createable    BOOLEAN     NOT NULL DEFAULT TRUE,
    updateable    BOOLEAN     NOT NULL DEFAULT TRUE,
    deletable     BOOLEAN     NOT NULL DEFAULT TRUE,
    queryable     BOOLEAN     NOT NULL DEFAULT TRUE,
    searchable    BOOLEAN     NOT NULL DEFAULT TRUE,
    replicateable BOOLEAN     NOT NULL DEFAULT TRUE,
    retrieveable  BOOLEAN     NOT NULL DEFAULT TRUE,
    undeletable   BOOLEAN     NOT NULL DEFAULT TRUE,
    triggerable   BOOLEAN     NOT NULL DEFAULT TRUE,
    is_tooling    BOOLEAN     NOT NULL DEFAULT FALSE,
    table_name    TEXT        NOT NULL
);

CREATE TABLE salesforce.sobject_field_definition
(
    id            SERIAL PRIMARY KEY,
    sobject_name  TEXT    NOT NULL REFERENCES salesforce.sobject_definition (name) ON DELETE CASCADE,
    name          TEXT    NOT NULL,
    column_name   TEXT    NOT NULL,
    label         TEXT    NOT NULL,
    field_type    TEXT    NOT NULL,
    soap_type     TEXT    NOT NULL,
    length        INT     NOT NULL DEFAULT 0,
    nillable      BOOLEAN NOT NULL DEFAULT TRUE,
    createable    BOOLEAN NOT NULL DEFAULT TRUE,
    updateable    BOOLEAN NOT NULL DEFAULT TRUE,
    filterable    BOOLEAN NOT NULL DEFAULT TRUE,
    sortable      BOOLEAN NOT NULL DEFAULT TRUE,
    custom        BOOLEAN NOT NULL DEFAULT FALSE,
    external_id   BOOLEAN NOT NULL DEFAULT FALSE,
    unique_field  BOOLEAN NOT NULL DEFAULT FALSE,
    name_field    BOOLEAN NOT NULL DEFAULT FALSE,
    reference_to  TEXT,
    UNIQUE (sobject_name, name)
);

-- =============================================================================
-- 3. STANDARD & CUSTOM SOBJECT RELATIONAL TABLES
-- =============================================================================
CREATE TABLE salesforce.account
(
    id                   VARCHAR(18) PRIMARY KEY DEFAULT generate_sf_id('001'),
    external_id__c       TEXT UNIQUE,
    name                 TEXT        NOT NULL,
    type                 TEXT,
    parent_id            VARCHAR(18),
    billing_street       TEXT,
    billing_city         TEXT,
    billing_state        TEXT,
    billing_postal_code  TEXT,
    billing_country      TEXT,
    shipping_street      TEXT,
    shipping_city        TEXT,
    shipping_state       TEXT,
    shipping_postal_code TEXT,
    shipping_country     TEXT,
    phone                TEXT,
    fax                  TEXT,
    account_number       TEXT,
    website              TEXT,
    industry             TEXT,
    annual_revenue       DOUBLE PRECISION,
    number_of_employees  INT,
    ownership            TEXT,
    ticker_symbol        TEXT,
    description          TEXT,
    rating               TEXT,
    site                 TEXT,
    owner_id             VARCHAR(18) NOT NULL    DEFAULT '005000000000001AAA',
    created_by_id        VARCHAR(18) NOT NULL    DEFAULT '005000000000001AAA',
    created_date         TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    last_modified_by_id  VARCHAR(18) NOT NULL    DEFAULT '005000000000001AAA',
    last_modified_date   TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    system_modstamp      TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    is_deleted           BOOLEAN     NOT NULL    DEFAULT FALSE
);

CREATE TABLE salesforce.contact
(
    id                  VARCHAR(18) PRIMARY KEY DEFAULT generate_sf_id('003'),
    external_id__c      TEXT UNIQUE,
    account_id          VARCHAR(18) REFERENCES salesforce.account (id) ON DELETE SET NULL,
    last_name           TEXT        NOT NULL,
    first_name          TEXT,
    salutation          TEXT,
    name                TEXT        NOT NULL,
    mailing_street      TEXT,
    mailing_city        TEXT,
    mailing_state       TEXT,
    mailing_postal_code TEXT,
    mailing_country     TEXT,
    phone               TEXT,
    fax                 TEXT,
    mobile_phone        TEXT,
    email               TEXT,
    title               TEXT,
    department          TEXT,
    lead_source         TEXT,
    description         TEXT,
    owner_id            VARCHAR(18) NOT NULL    DEFAULT '005000000000001AAA',
    created_by_id       VARCHAR(18) NOT NULL    DEFAULT '005000000000001AAA',
    created_date        TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    last_modified_by_id VARCHAR(18) NOT NULL    DEFAULT '005000000000001AAA',
    last_modified_date  TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    system_modstamp     TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    is_deleted          BOOLEAN     NOT NULL    DEFAULT FALSE
);

CREATE TABLE salesforce.lead
(
    id                   VARCHAR(18) PRIMARY KEY DEFAULT generate_sf_id('00Q'),
    external_id__c       TEXT UNIQUE,
    last_name            TEXT        NOT NULL,
    first_name           TEXT,
    salutation           TEXT,
    name                 TEXT        NOT NULL,
    title                TEXT,
    company              TEXT        NOT NULL,
    street               TEXT,
    city                 TEXT,
    state                TEXT,
    postal_code          TEXT,
    country              TEXT,
    phone                TEXT,
    mobile_phone         TEXT,
    email                TEXT,
    website              TEXT,
    description          TEXT,
    lead_source          TEXT,
    status               TEXT        NOT NULL    DEFAULT 'Open - Not Contacted',
    industry             TEXT,
    rating               TEXT,
    annual_revenue       DOUBLE PRECISION,
    number_of_employees  INT,
    is_converted         BOOLEAN     NOT NULL    DEFAULT FALSE,
    converted_account_id VARCHAR(18),
    converted_contact_id VARCHAR(18),
    owner_id             VARCHAR(18) NOT NULL    DEFAULT '005000000000001AAA',
    created_by_id        VARCHAR(18) NOT NULL    DEFAULT '005000000000001AAA',
    created_date         TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    last_modified_by_id  VARCHAR(18) NOT NULL    DEFAULT '005000000000001AAA',
    last_modified_date   TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    system_modstamp      TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    is_deleted           BOOLEAN     NOT NULL    DEFAULT FALSE
);

CREATE TABLE salesforce.case
(
    id                            VARCHAR(18) PRIMARY KEY DEFAULT generate_sf_id('500'),
    external_id__c                TEXT UNIQUE,
    case_number                   TEXT        NOT NULL    DEFAULT lpad((floor(random() * 99999999)::int)::text, 8, '0'),
    additionalinfo__c             TEXT,
    sub_type__c                   TEXT,
    type__c                       TEXT,
    type_name                     TEXT,
    origin                        TEXT,
    record_type_id                VARCHAR(18),
    parent_id                     VARCHAR(18),
    population_impacted__c        TEXT,
    closed_date                   TIMESTAMPTZ,
    created_date                  TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    weblinks__c                   TEXT,
    root_cause__c                 TEXT,
    impact__c                     TEXT,
    vip__c                        BOOLEAN,
    priority                      TEXT        NOT NULL    DEFAULT 'Medium',
    account_id                    VARCHAR(18),
    contact_id                    VARCHAR(18),
    recurrence__c                 TEXT,
    human_factor__c               TEXT,
    technical_system__c           TEXT,
    contact_conclusion_level_2__c TEXT,
    contact_conclusion_level_1__c TEXT,
    contact_action_level_3__c     TEXT,
    contact_action_level_2__c     TEXT,
    contact_action_level_1__c     TEXT,
    contact_topic_level_3__c      TEXT,
    contact_topic_level_2__c      TEXT,
    contact_topic_level_1__c      TEXT,
    components__c                 TEXT,
    steps_to_reproduce__c         TEXT,
    expected_result__c            TEXT,
    actual_result__c              TEXT,
    description                   TEXT,
    owner_id                      VARCHAR(18)             DEFAULT '005000000000001AAA',
    sub_status_level_3__c         TEXT,
    sub_status_level_2__c         TEXT,
    sub_status__c                 TEXT,
    status                        TEXT                    DEFAULT 'New',
    created_by_id                 VARCHAR(18)             DEFAULT '005000000000001AAA',
    subject                       TEXT,
    send_communication__c         BOOLEAN     NOT NULL    DEFAULT FALSE,
    last_modified_date            TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    system_modstamp               TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    is_deleted                    BOOLEAN     NOT NULL    DEFAULT FALSE
);

CREATE TABLE salesforce.custom_object__c
(
    id                  VARCHAR(18) PRIMARY KEY DEFAULT generate_sf_id('a00'),
    external_id__c      TEXT UNIQUE,
    name                TEXT        NOT NULL,
    field_to_update__c  TEXT,
    description__c      TEXT,
    status__c           TEXT,
    owner_id            VARCHAR(18) NOT NULL    DEFAULT '005000000000001AAA',
    created_by_id       VARCHAR(18) NOT NULL    DEFAULT '005000000000001AAA',
    created_date        TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    last_modified_by_id VARCHAR(18) NOT NULL    DEFAULT '005000000000001AAA',
    last_modified_date  TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    system_modstamp     TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    is_deleted          BOOLEAN     NOT NULL    DEFAULT FALSE
);

CREATE TABLE salesforce.content_document
(
    id           VARCHAR(18) PRIMARY KEY DEFAULT generate_sf_id('069'),
    type         TEXT,
    title        TEXT,
    created_date TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    is_deleted   BOOLEAN     NOT NULL    DEFAULT FALSE
);

CREATE TABLE salesforce.content_version
(
    id                        VARCHAR(18) PRIMARY KEY DEFAULT generate_sf_id('068'),
    content_document_id       VARCHAR(18) NOT NULL REFERENCES salesforce.content_document (id) ON DELETE CASCADE,
    is_latest                 BOOLEAN                 DEFAULT TRUE,
    content_url               TEXT,
    content_body_id           VARCHAR(18),
    version_number            TEXT        NOT NULL    DEFAULT '1',
    title                     TEXT        NOT NULL,
    description               TEXT,
    reason_for_change         TEXT,
    sharing_option            TEXT,
    sharing_privacy           TEXT,
    path_on_client            TEXT        NOT NULL,
    rating_count              INT                     DEFAULT 0,
    is_deleted                BOOLEAN                 DEFAULT FALSE,
    content_modified_date     TIMESTAMPTZ,
    content_modified_by_id    VARCHAR(18),
    positive_rating_count     INT                     DEFAULT 0,
    negative_rating_count     INT                     DEFAULT 0,
    featured_content_boost    INT,
    featured_content_date     TIMESTAMPTZ,
    owner_id                  VARCHAR(18)             DEFAULT '005000000000001AAA',
    created_by_id             VARCHAR(18)             DEFAULT '005000000000001AAA',
    created_date              TIMESTAMPTZ             DEFAULT CURRENT_TIMESTAMP,
    last_modified_by_id       VARCHAR(18),
    last_modified_date        TIMESTAMPTZ             DEFAULT CURRENT_TIMESTAMP,
    system_modstamp           TIMESTAMPTZ             DEFAULT CURRENT_TIMESTAMP,
    tag_csv                   TEXT,
    file_type                 TEXT,
    publish_status            TEXT                    DEFAULT 'P',
    version_data              TEXT        NOT NULL,
    content_size              INT,
    file_extension            TEXT,
    first_publish_location_id VARCHAR(18),
    origin                    TEXT                    DEFAULT 'C',
    network_id                VARCHAR(18),
    content_location          TEXT                    DEFAULT 'S',
    text_preview              TEXT,
    external_document_info1   TEXT,
    external_document_info2   TEXT,
    external_data_source_id   VARCHAR(18),
    checksum                  TEXT,
    is_major_version          BOOLEAN                 DEFAULT TRUE,
    is_encrypted              BOOLEAN                 DEFAULT FALSE,
    is_asset_enabled          BOOLEAN                 DEFAULT FALSE,
    version_data_url          TEXT
);

CREATE TABLE salesforce.content_document_link
(
    id                  VARCHAR(18) PRIMARY KEY DEFAULT generate_sf_id('06A'),
    content_document_id VARCHAR(18) NOT NULL REFERENCES salesforce.content_document (id) ON DELETE CASCADE,
    linked_entity_id    VARCHAR(18) NOT NULL,
    share_type          TEXT        NOT NULL    DEFAULT 'V',
    visibility          TEXT        NOT NULL    DEFAULT 'AllUsers',
    is_deleted          BOOLEAN     NOT NULL    DEFAULT FALSE
);

-- =============================================================================
-- 4. SOQL PAGINATION CURSOR TABLE
-- =============================================================================
CREATE TABLE salesforce.query_cursor
(
    id              VARCHAR(32) PRIMARY KEY DEFAULT generate_ulid(),
    sobject_name    TEXT        NOT NULL,
    soql_query      TEXT        NOT NULL,
    include_deleted BOOLEAN     NOT NULL    DEFAULT FALSE,
    offset_value    INT         NOT NULL    DEFAULT 0,
    batch_size      INT         NOT NULL    DEFAULT 2000,
    total_size      INT         NOT NULL    DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- 5. BULK API 2.0 TABLES
-- =============================================================================
CREATE TABLE salesforce.bulk_ingest_job
(
    id                       VARCHAR(18) PRIMARY KEY DEFAULT generate_sf_id('750'),
    operation                TEXT        NOT NULL,
    object_name              TEXT        NOT NULL,
    external_id_field_name   TEXT,
    content_type             TEXT        NOT NULL    DEFAULT 'CSV',
    column_delimiter         TEXT        NOT NULL    DEFAULT 'COMMA',
    line_ending              TEXT        NOT NULL    DEFAULT 'LF',
    state                    TEXT        NOT NULL    DEFAULT 'Open',
    concurrency_mode         TEXT        NOT NULL    DEFAULT 'Parallel',
    api_version              TEXT        NOT NULL    DEFAULT '61.0',
    number_records_processed INT         NOT NULL    DEFAULT 0,
    number_records_failed    INT         NOT NULL    DEFAULT 0,
    retries                  INT         NOT NULL    DEFAULT 0,
    total_processing_time    BIGINT      NOT NULL    DEFAULT 0,
    created_by_id            VARCHAR(18) NOT NULL    DEFAULT '005000000000001AAA',
    created_date             TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    system_modstamp          TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE salesforce.bulk_ingest_batch
(
    id           VARCHAR(18) PRIMARY KEY DEFAULT generate_sf_id('751'),
    job_id       VARCHAR(18) NOT NULL REFERENCES salesforce.bulk_ingest_job (id) ON DELETE CASCADE,
    csv_content  TEXT        NOT NULL,
    created_date TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE salesforce.bulk_ingest_result
(
    id           BIGSERIAL PRIMARY KEY,
    job_id       VARCHAR(18) NOT NULL REFERENCES salesforce.bulk_ingest_job (id) ON DELETE CASCADE,
    sf_id        VARCHAR(18) NOT NULL DEFAULT '',
    sf_created   BOOLEAN     NOT NULL DEFAULT FALSE,
    sf_error     TEXT        NOT NULL DEFAULT '',
    is_success   BOOLEAN     NOT NULL DEFAULT TRUE,
    row_csv      TEXT        NOT NULL,
    created_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE salesforce.bulk_query_job
(
    id                       VARCHAR(18) PRIMARY KEY DEFAULT generate_sf_id('750'),
    operation                TEXT        NOT NULL    DEFAULT 'query',
    query_text               TEXT        NOT NULL,
    object_name              TEXT        NOT NULL,
    content_type             TEXT        NOT NULL    DEFAULT 'CSV',
    column_delimiter         TEXT        NOT NULL    DEFAULT 'COMMA',
    line_ending              TEXT        NOT NULL    DEFAULT 'LF',
    state                    TEXT        NOT NULL    DEFAULT 'JobComplete',
    concurrency_mode         TEXT        NOT NULL    DEFAULT 'Parallel',
    api_version              TEXT        NOT NULL    DEFAULT '61.0',
    number_records_processed INT         NOT NULL    DEFAULT 0,
    retries                  INT         NOT NULL    DEFAULT 0,
    total_processing_time    BIGINT      NOT NULL    DEFAULT 0,
    created_by_id            VARCHAR(18) NOT NULL    DEFAULT '005000000000001AAA',
    created_date             TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    system_modstamp          TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP
);

-- =============================================================================
-- 6. TOOLING API TABLES
-- =============================================================================
CREATE TABLE salesforce.apex_class
(
    id                      VARCHAR(18) PRIMARY KEY DEFAULT generate_sf_id('01p'),
    name                    TEXT UNIQUE      NOT NULL,
    body                    TEXT             NOT NULL,
    api_version             DOUBLE PRECISION NOT NULL DEFAULT 61.0,
    status                  TEXT             NOT NULL DEFAULT 'Active',
    is_valid                BOOLEAN          NOT NULL DEFAULT TRUE,
    length_without_comments INT              NOT NULL DEFAULT 0,
    namespace_prefix        TEXT,
    created_by_id           VARCHAR(18)      NOT NULL DEFAULT '005000000000001AAA',
    created_date            TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_modified_by_id     VARCHAR(18)      NOT NULL DEFAULT '005000000000001AAA',
    last_modified_date      TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    system_modstamp         TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE salesforce.apex_trigger
(
    id                  VARCHAR(18) PRIMARY KEY DEFAULT generate_sf_id('01q'),
    name                TEXT UNIQUE      NOT NULL,
    table_enum_or_id    TEXT             NOT NULL,
    body                TEXT             NOT NULL,
    api_version         DOUBLE PRECISION NOT NULL DEFAULT 61.0,
    status              TEXT             NOT NULL DEFAULT 'Active',
    is_valid            BOOLEAN          NOT NULL DEFAULT TRUE,
    usage_before_insert BOOLEAN          NOT NULL DEFAULT FALSE,
    usage_after_insert  BOOLEAN          NOT NULL DEFAULT FALSE,
    usage_before_update BOOLEAN          NOT NULL DEFAULT FALSE,
    usage_after_update  BOOLEAN          NOT NULL DEFAULT FALSE,
    created_by_id       VARCHAR(18)      NOT NULL DEFAULT '005000000000001AAA',
    created_date        TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_modified_date  TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE salesforce.custom_field
(
    id                 VARCHAR(18) PRIMARY KEY DEFAULT generate_sf_id('00N'),
    developer_name     TEXT        NOT NULL,
    table_enum_or_id   TEXT        NOT NULL,
    full_name          TEXT UNIQUE NOT NULL,
    label              TEXT        NOT NULL,
    data_type          TEXT        NOT NULL    DEFAULT 'Text',
    length             INT         NOT NULL    DEFAULT 255,
    created_date       TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    last_modified_date TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE salesforce.validation_rule
(
    id                      VARCHAR(18) PRIMARY KEY DEFAULT generate_sf_id('03d'),
    validation_name         TEXT        NOT NULL,
    entity_definition_id    TEXT        NOT NULL,
    active                  BOOLEAN     NOT NULL    DEFAULT TRUE,
    description             TEXT,
    error_condition_formula TEXT        NOT NULL,
    error_message           TEXT        NOT NULL,
    created_date            TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    last_modified_date      TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE salesforce.apex_test_run
(
    id                VARCHAR(18) PRIMARY KEY DEFAULT generate_sf_id('707'),
    class_names       TEXT        NOT NULL,
    test_level        TEXT        NOT NULL    DEFAULT 'RunSpecifiedTests',
    status            TEXT        NOT NULL    DEFAULT 'Completed',
    methods_enqueued  INT         NOT NULL    DEFAULT 1,
    methods_completed INT         NOT NULL    DEFAULT 1,
    methods_failed    INT         NOT NULL    DEFAULT 0,
    created_date      TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    completed_date    TIMESTAMPTZ
);

-- =============================================================================
-- 7. APEX REST & EVENT BUS (COMETD / PLATFORM EVENTS) TABLES
-- =============================================================================
CREATE TABLE salesforce.apex_rest_route
(
    id              VARCHAR(18) PRIMARY KEY DEFAULT generate_sf_id('01p'),
    url_mapping     TEXT        NOT NULL,
    http_method     TEXT        NOT NULL    DEFAULT 'GET',
    class_name      TEXT        NOT NULL,
    response_status INT         NOT NULL    DEFAULT 200,
    response_body   TEXT        NOT NULL,
    is_active       BOOLEAN     NOT NULL    DEFAULT TRUE,
    created_date    TIMESTAMPTZ NOT NULL    DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (url_mapping, http_method)
);

CREATE TABLE salesforce.platform_event
(
    id            VARCHAR(18) PRIMARY KEY DEFAULT generate_sf_id('e00'),
    replay_id     BIGSERIAL UNIQUE NOT NULL,
    channel_name  TEXT             NOT NULL,
    event_type    TEXT             NOT NULL,
    payload       TEXT             NOT NULL,
    created_by_id VARCHAR(18)      NOT NULL DEFAULT '005000000000001AAA',
    created_date  TIMESTAMPTZ      NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE salesforce.cometd_client
(
    client_id    VARCHAR(36) PRIMARY KEY,
    is_connected BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE salesforce.cometd_subscription
(
    id             BIGSERIAL PRIMARY KEY,
    client_id      VARCHAR(36) NOT NULL REFERENCES salesforce.cometd_client (client_id) ON DELETE CASCADE,
    channel_name   TEXT        NOT NULL,
    last_replay_id BIGINT      NOT NULL DEFAULT -1,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (client_id, channel_name)
);

-- +goose Down
DROP SCHEMA IF EXISTS salesforce CASCADE;
