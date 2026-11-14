# AGENTS.md — Salesforce API Emulator (`sem`)

## Overview

`sem` (`github.com/faber-numeris/sem`) is a local Salesforce Platform API Emulator designed for integration testing, local development, and CI pipelines. It emulates the standard Salesforce API surface backed by normalized PostgreSQL tables.

---

## 1. Hexagonal Architecture & Directory Structure

This project strictly follows the **Hexagonal Architecture (Ports and Adapters)** pattern inspired by `../beholder`:

```text
.
├── api/                                         # Generated OpenAPI types & server interfaces (DO NOT EDIT)
├── bruno/                                       # Bruno OpenCollection API catalog & Integration Tests
│   ├── environments/
│   │   └── Local development server.yml         # Local emulator environment variables
│   └── Integration Tests/                       # Automated Bruno end-to-end integration test suite
├── cmd/
│   └── main.go                                  # Application entrypoint
├── internal/
│   ├── adapters/
│   │   ├── inbound/
│   │   │   └── httpapi/                         # Inbound HTTP adapter (Chi router & handlers)
│   │   └── outbound/
│   │       └── postgres/                        # Outbound PostgreSQL adapter
│   │           ├── gen/                         # sqlc-generated persistence code (DO NOT EDIT)
│   │           ├── migrations/                  # Goose SQL migrations & seed data
│   │           └── queries/                     # sqlc SQL query definitions
│   ├── app/
│   │   └── bootstrap/                           # Dependency injection & server lifecycle wiring
│   ├── core/
│   │   ├── domain/                              # Pure domain entities, value objects, and errors
│   │   └── services/                            # Application business logic implementing inbound ports
│   ├── infrastructure/
│   │   ├── config/                              # Environment & YAML configuration loading
│   │   └── postgres/                            # pgxpool connection management & AbstractPool interface
│   ├── mocks/                                   # mockery-generated test doubles (DO NOT EDIT)
│   ├── platform/
│   │   └── util/                                # Shared platform helpers (e.g., Must)
│   └── ports/
│       ├── inbound/                             # Inbound service interfaces (EmulatorService, etc.)
│       └── outbound/                            # Outbound repository interfaces (Repository, etc.)
├── .mockery.yaml                                # Mockery configuration
├── config.yaml                                  # Default local service & database configuration
├── docker-compose.yaml                          # Local PostgreSQL, Goose migration, and API containers
├── Dockerfile                                   # Multi-stage Go container build
├── justfile                                     # Build, codegen, migration, and test automation recipes
└── sqlc.yaml                                    # sqlc configuration
```

### Dependency Rules
- **`internal/core/domain` & `internal/core/services`** MUST NOT import any inbound/outbound adapter (`httpapi`, `postgres/gen`, `pgx`, `chi`) or infrastructure package. Core packages depend **only** on `internal/ports/` and the Go standard library.
- **`GOWORK=off`**: Because the parent workspace (`../go.work`) does not include this module, always build, test, and run Go commands with `GOWORK=off` (automatically exported in `justfile`).

---

## 2. Database Migrations (`goose`) & Persistence Layer (`sqlc`)

### Relational Schema & Migrations
- All standard Salesforce entities and sObjects (`Account`, `Contact`, `Lead`, `Opportunity`, `Case`, `Custom_Object__c`, `ContentDocument`, `ContentVersion`, `ContentDocumentLink`, OAuth tables, Bulk v2 jobs/batches/results, Tooling metadata, Apex REST routes, and CometD/Platform Events) are modeled as **normalized relational PostgreSQL tables** — never use a generic JSONB blob table for sObjects.
- Migrations live in `internal/adapters/outbound/postgres/migrations/` and are managed with **goose** (`salesforce_migrations` version table):
  - Apply migrations: `just migration-up`
  - Roll back last migration: `just migration-down`
  - Check migration status: `just migration-status`
  - Create a new migration: `just migration-new <name>`

### `sqlc` Persistence Rules
- All SQL access goes through **sqlc**-generated code in `internal/adapters/outbound/postgres/gen/`.
- **Never** write hardcoded SQL strings in Go repository/adapter files (`internal/adapters/outbound/postgres/*.go`).
- **Never** hand-edit files under `internal/adapters/outbound/postgres/gen/`.
- Workflow for adding or modifying queries:
  1. Add or edit the SQL query in `internal/adapters/outbound/postgres/queries/*.sql` using `sqlc.arg` / `sqlc.narg`.
  2. Run `just generate-sqlc` to regenerate `internal/adapters/outbound/postgres/gen/`.
  3. Call the generated method via the `gen.Querier` interface in `internal/adapters/outbound/postgres/*.go`.
- Nullable `timestamptz` columns must use the `*time.Time` pointer override in `sqlc.yaml`.

---

## 3. Unit Testing, Mocking (`mockery`), and 100% Coverage Policy

### Strict Testing Directives
1. **100% Unit Test Coverage is Mandatory**:
   - Every handwritten package under `./internal/...` (excluding generated packages `/gen`, `/mocks`, `/generated`) **must maintain 100.0% statement coverage**.
   - Run `just test-coverage` to verify. The command fails if coverage drops below `100.0%`.
2. **No `testcontainers` and No `sqlmock`**:
   - Never use `testcontainers` or `DATA-DOG/go-sqlmock` in unit tests.
3. **Mocks via `mockery` Only**:
   - All mocks are generated by **mockery** (`v2.53.3`) into `internal/mocks/` based on `.mockery.yaml`.
   - Regenerate mocks strictly via `just generate-mocks` — never hand-edit `internal/mocks/`.
   - Unit tests for outbound PostgreSQL repositories inject `mocks.NewMockQuerier(t)` / `mocks.NewMockDBTX(t)`.
   - Unit tests for core services inject `mocks.NewMockRepository(t)`.
   - Unit tests for inbound HTTP handlers inject `mocks.NewMockEmulatorService(t)`.
   - Unit tests for infrastructure & bootstrap inject `mocks.NewMockAbstractPool(t)` and `mocks.NewMockIDatabaseConfig(t)`.

---

## 4. Local Environment & Docker Compose

- Run the entire stack locally (PostgreSQL 14 + Goose migrations + `sem` API on `http://localhost:8080`) using Docker Compose:
  ```bash
  just up
  ```
- Stop the Docker Compose stack:
  ```bash
  just down
  ```
- Run the API directly on the host against a local PostgreSQL instance:
  ```bash
  just migration-up
  just run
  ```

---

## 5. Bruno Integration Tests (`bruno/`)

- Integration tests are authored using **Bruno** in the `bruno/Integration Tests/` folder and configured against the `Local development server` environment (`bruno/environments/Local development server.yml`).
- Whenever endpoints are added or modified, add or update the corresponding `.yml` test spec in `bruno/Integration Tests/`.
- Run the Bruno integration test suite against a running local instance (`http://localhost:8080`):
  ```bash
  just test-bruno
  ```

---

## 6. Emulated Standard Salesforce API Endpoints

The emulator implements the following 7 endpoint groups (where `<API_VERSION>` supports e.g. `v60.0`, `v61.0`, `v66.0` and `<SOBJECT>` supports `Account`, `Contact`, `Lead`, `Opportunity`, `Case`, `Custom_Object__c`, `ContentDocument`, `ContentVersion`, `ContentDocumentLink`, and Platform Events `<Event>__e`):

### 1. OAuth 2.0 & Authentication Endpoints
- `POST /services/oauth2/token` — Token Endpoint (`client_credentials`, `authorization_code`, `password`, `refresh_token`, `urn:ietf:params:oauth:grant-type:jwt-bearer`)
- `GET /services/oauth2/authorize` — Authorization Code & Implicit Flow (`response_type=code|token`, `client_id`, `redirect_uri`, `scope`, `state`)
- `POST /services/oauth2/revoke?token=<ACCESS_OR_REFRESH_TOKEN>` — Token Revocation
- `POST /services/oauth2/introspect` — Token Introspection (`token`, `token_type_hint`, `client_id`, `client_secret`)
- `GET|POST /services/oauth2/userinfo` — OpenID Connect UserInfo (`Authorization: Bearer <ACCESS_TOKEN>`)

### 2. REST API Core Endpoints (CRUD & Metadata)
- `GET /services/data/` — API Versions Listing
- `GET /services/data/<API_VERSION>/` — Available Resources for Version
- `GET /services/data/<API_VERSION>/limits` — Organization Limits
- `GET /services/data/<API_VERSION>/sobjects/` — Global Describe
- `GET /services/data/<API_VERSION>/sobjects/<SOBJECT>` — sObject Basic Information & Recent Items
- `GET /services/data/<API_VERSION>/sobjects/<SOBJECT>/describe` — sObject Describe & Field Metadata
- `POST /services/data/<API_VERSION>/sobjects/<SOBJECT>/` — Create Record
- `GET /services/data/<API_VERSION>/sobjects/<SOBJECT>/<RECORD_ID>?fields=Id,Name,CreatedDate` — Retrieve Record (with optional field projection)
- `GET /services/data/<API_VERSION>/sobjects/ContentVersion/<RECORD_ID>/VersionData` — Retrieve ContentVersion Binary Payload
- `PATCH /services/data/<API_VERSION>/sobjects/<SOBJECT>/<RECORD_ID>` — Update Record
- `PATCH /services/data/<API_VERSION>/sobjects/<SOBJECT>/<EXTERNAL_ID_FIELD>/<EXTERNAL_ID_VALUE>` — Upsert Record by External ID
- `DELETE /services/data/<API_VERSION>/sobjects/<SOBJECT>/<RECORD_ID>` — Delete Record (Soft Delete)

### 3. Query & Search (SOQL / SOSL)
- `GET /services/data/<API_VERSION>/query/?q=<URL_ENCODED_SOQL>` — Execute SOQL Query (respects `Sforce-Query-Options: batchSize=N`)
- `GET /services/data/<API_VERSION>/query/<NEXT_RECORDS_IDENTIFIER>` — Pagination (Next Records Query)
- `GET /services/data/<API_VERSION>/queryAll/?q=<URL_ENCODED_SOQL>` — Execute SOQL Query All (includes soft-deleted records)
- `GET /services/data/<API_VERSION>/search/?q=<URL_ENCODED_SOSL>` — Execute SOSL Search (`FIND {term} IN ... RETURNING ...`)

### 4. Bulk API 2.0 (High-Volume Data Operations)
- `POST /services/data/<API_VERSION>/jobs/ingest/` — Create Bulk Ingest Job (`insert`, `update`, `upsert`, `delete`)
- `GET /services/data/<API_VERSION>/jobs/ingest/` — List Bulk Ingest Jobs
- `PUT /services/data/<API_VERSION>/jobs/ingest/<JOB_ID>/batches` — Upload CSV Data Batch (`Content-Type: text/csv`)
- `PATCH /services/data/<API_VERSION>/jobs/ingest/<JOB_ID>/` — Close (`UploadComplete`) or Abort (`Aborted`) Ingest Job
- `GET /services/data/<API_VERSION>/jobs/ingest/<JOB_ID>/` — Check Bulk Ingest Job Status
- `GET /services/data/<API_VERSION>/jobs/ingest/<JOB_ID>/successfulResults/` — Retrieve Successful Results CSV
- `GET /services/data/<API_VERSION>/jobs/ingest/<JOB_ID>/failedResults/` — Retrieve Failed Results CSV
- `POST /services/data/<API_VERSION>/jobs/query/` — Create Bulk Query Job
- `GET /services/data/<API_VERSION>/jobs/query/<JOB_ID>` — Check Bulk Query Job Status
- `GET /services/data/<API_VERSION>/jobs/query/<JOB_ID>/results?maxRecords=10000` — Retrieve Bulk Query Results CSV

### 5. Tooling API (Developer & Deployment Tooling)
- `GET /services/data/<API_VERSION>/tooling/query/?q=<URL_ENCODED_SOQL>` — Tooling SOQL Query (`ApexClass`, `ApexTrigger`, `CustomField`, `ValidationRule`)
- `GET /services/data/<API_VERSION>/tooling/executeAnonymous/?anonymousBody=<URL_ENCODED_APEX>` — Execute Anonymous Apex
- `GET /services/data/<API_VERSION>/tooling/sobjects/` — Tooling Global Describe
- `GET /services/data/<API_VERSION>/tooling/sobjects/<TOOLING_OBJECT>/describe` — Tooling sObject Describe
- `POST /services/data/<API_VERSION>/tooling/runTestsAsynchronous/` — Run Apex Unit Tests Asynchronous

### 6. Apex REST API & Event Bus Conventions
- `ANY /services/apexrest/<URL_MAPPING>` — Custom Apex REST Dispatch (e.g. `/services/apexrest/v1/integrations/status`)
- `POST /cometd/<API_VERSION>/` — Bayeux/CometD Streaming API (`/meta/handshake`, `/meta/subscribe`, `/meta/unsubscribe`, `/meta/connect`, `/meta/disconnect`, `/event/<CUSTOM_EVENT__e>`, `/data/ChangeEvents`)

### 7. Composite & Batch API (Optimized Multi-Requests)
- `POST /services/data/<API_VERSION>/composite` — Sequential Sub-Requests with `@{referenceId.field}` Interpolation
- `POST /services/data/<API_VERSION>/composite/batch` — Independent Multi-Request Batch Execution
