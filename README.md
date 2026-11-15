# Salesforce Emulator (`sem`)

A local emulator for the Salesforce Platform APIs, backed by PostgreSQL. Designed for integration testing, local development, and CI pipelines.

## Architecture

This project strictly follows the **Hexagonal Architecture (Ports and Adapters)** pattern to ensure a clean separation between core business logic and infrastructure concerns.

- Core packages (`internal/core/domain` and `internal/core/services`) depend **only** on `internal/ports/` and the Go standard library—never on HTTP, database drivers, or generated adapter code.
- See [docs/hexagonal-architecture.md](docs/hexagonal-architecture.md) for an overview of the architecture, layers, and dependency rules.

## Project Structure

```text
.
├── .github/
│   └── workflows/
│       └── docker-publish.yml                   # CI/CD: tests & publishes image to ghcr.io/faber-numeris/sem
├── api/                                         # Generated OpenAPI types & server interfaces (DO NOT EDIT)
├── bruno/                                       # Bruno OpenCollection API catalog & Integration Tests
│   ├── environments/
│   │   └── Local development server.yml         # Local emulator environment variables
│   └── Integration Tests/                       # Automated Bruno end-to-end integration test suite
├── cmd/
│   └── main.go                                  # Application entrypoint
├── docs/
│   └── hexagonal-architecture.md                # Short explanation of Hexagonal Architecture in sem
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

### What We Expect in Each Directory

- **`cmd/`**: Main application entrypoint; delegates directly to the bootstrap package.
- **`internal/core/domain/`**: Pure domain models, value objects, and domain-level errors with zero external dependencies.
- **`internal/core/services/`**: Business logic implementing the inbound service interfaces (`internal/ports/inbound`).
- **`internal/ports/`**: Explicit interface contracts separating the core from external adapters (`inbound` for driving ports, `outbound` for driven ports).
- **`internal/adapters/inbound/httpapi/`**: HTTP transport layer routing Salesforce REST, OAuth2, Bulk v2, Tooling, Apex REST, CometD, and Composite requests to the core service.
- **`internal/adapters/outbound/postgres/`**: Relational persistence implementation backed by PostgreSQL, Goose migrations (`migrations/`), SQL query definitions (`queries/`), and `sqlc`-generated code (`gen/`).
- **`internal/infrastructure/`**: Cross-cutting technical setup such as configuration parsing (`config/`) and PostgreSQL connection pooling (`postgres/`).
- **`internal/app/bootstrap/`**: Application wiring (Dependency Injection) that connects infrastructure, outbound adapters, core services, and inbound adapters.
- **`internal/mocks/`**: Generated `mockery` test doubles used to maintain 100% unit test coverage across handwritten packages without external containers.
- **`bruno/`**: End-to-end integration test suite executed against the running emulator.

## Technical Stack

- **Language:** Go 1.26
- **Database:** PostgreSQL
- **Code Generation & Persistence:** `sqlc`, `goose` (migrations), `mockery` (unit test mocks)
- **Task Runner & Containers:** `justfile`, Docker & Docker Compose
- **Integration Testing:** Bruno CLI
