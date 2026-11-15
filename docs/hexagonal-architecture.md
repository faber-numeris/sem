# Hexagonal Architecture (Ports and Adapters)

`sem` follows a **Hexagonal Architecture** (also known as **Ports and Adapters**) to keep core Salesforce emulation logic decoupled from transport mechanisms, databases, and external frameworks.

## Core Principle

The architecture enforces a strict inward dependency rule: **inner layers never depend on outer layers**. Core domain models and application services know nothing about HTTP routers, SQL queries, or PostgreSQL drivers—they interact with the outside world strictly through **Ports** (Go interfaces).

## Layers

1. **Core Domain (`internal/core/domain`)**
   - Contains pure domain entities, value objects, and domain errors.
   - Has zero dependencies on adapters, infrastructure packages, or external frameworks.

2. **Ports (`internal/ports`)**
   - Defines the boundaries of the application as Go interfaces:
     - **Inbound Ports (`internal/ports/inbound`)**: Interfaces exposed by the core application (e.g., `EmulatorService`) and called by driving adapters.
     - **Outbound Ports (`internal/ports/outbound`)**: Interfaces required by the core application (e.g., `Repository`) and implemented by driven adapters.

3. **Core Services (`internal/core/services`)**
   - Implements the inbound ports and orchestrates application use cases (OAuth 2.0 flows, sObject CRUD, SOQL/SOSL execution, Bulk API 2.0, Tooling API, Apex REST, CometD, and Composite APIs).
   - Depends only on `internal/core/domain`, `internal/ports`, and the Go standard library.

4. **Adapters (`internal/adapters`)**
   - Translates between external protocols/storage and the core ports:
     - **Inbound Adapters (`internal/adapters/inbound/httpapi`)**: Driving adapter that handles HTTP requests via Chi router and delegates to `inbound.EmulatorService`.
     - **Outbound Adapters (`internal/adapters/outbound/postgres`)**: Driven adapter that implements `outbound.Repository` using PostgreSQL, `pgx`, and `sqlc`-generated type-safe queries.

5. **Infrastructure & Bootstrap (`internal/infrastructure`, `internal/app/bootstrap`, `cmd`)**
   - Manages configuration loading, database connection pooling (`pgxpool`), and dependency injection wiring at startup.
