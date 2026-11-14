# The parent go.work does not include this module; build it standalone.
export GOWORK := "off"

pg_user := env('PG_USER', 'postgres')
pg_password := env('PG_PASSWORD', 'postgres')
pg_host := env('PG_HOST', 'localhost')
pg_port := env('PG_PORT', '5432')
pg_db := env('PG_DB', 'postgres')
postgres_conn := "postgres://" + pg_user + ":" + pg_password + "@" + pg_host + ":" + pg_port + "/" + pg_db + "?sslmode=disable"

migration_dir := "internal/adapters/outbound/postgres/migrations"
goose := "go run github.com/pressly/goose/v3/cmd/goose@v3.27.0 -table salesforce_migrations -dir " + migration_dir

# List available recipes
default:
    @just --list

migration-new name="migration":
    {{goose}} create "{{name}}" sql

migration-down:
    {{goose}} postgres "{{postgres_conn}}" down

migrate-all-down: migration-down

migration-up-by-one:
    {{goose}} postgres "{{postgres_conn}}" up-by-one -allow-missing

migration-up:
    {{goose}} postgres "{{postgres_conn}}" up -allow-missing

migrate-all-up: migration-up

migration-status:
    {{goose}} postgres "{{postgres_conn}}" status

generate-sqlc:
    rm -rf ./internal/adapters/outbound/postgres/gen
    go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.29.0 generate

sqlc: generate-sqlc

generate-mocks:
    rm -rf ./internal/mocks
    go run github.com/vektra/mockery/v2@v2.53.3 --config .mockery.yaml

generate-mappers:
    go run github.com/jmattheis/goverter/cmd/goverter@v1.8.0 gen ./internal/platform/mapper

goverter: generate-mappers

# Build with the go.work workspace disabled
build:
    GOWORK=off go build ./...

# Run the API locally
run:
    GOWORK=off go run ./cmd

# Start local environment (PostgreSQL, migrations, and API) via Docker Compose
up:
    docker compose up -d --build

# Stop local Docker Compose environment
down:
    docker compose down

test:
    GOWORK=off go test ./... -count=1

test-coverage:
    #!/usr/bin/env bash
    set -euo pipefail
    pkgs=$(GOWORK=off go list ./internal/... | grep -v -E '/gen$|/mocks$|/generated$')
    GOWORK=off go test -coverprofile=coverage.out -count=1 $pkgs
    go tool cover -func=coverage.out | tail -n 1
    go tool cover -func=coverage.out | tail -n 1 | awk '{print $3}' | sed 's/%//' | awk '{if ($1 < 100) exit 1}'

# Run Bruno integration tests against local running instance
test-bruno:
    cd bruno && npx -y @usebruno/cli@4.3.0 run "Integration Tests" --env "Local development server"

# Convert the Postman collection into a Bruno OpenCollection (YAML) folder
postman-to-bruno source="Salesforce Platform APIs.postman_collection.json" output="bruno" name="Salesforce Platform APIs":
    SRC="{{source}}" OUT="{{output}}" NAME="{{name}}" \
    npx -y -p @usebruno/cli@4.3.0 -c 'NODE_PATH="$(dirname "$(dirname "$(command -v bru)")")" node tools/postman-to-bruno.js "$SRC" "$OUT" "$NAME"'

# Convert the Postman collection into an OpenAPI 3 specification (YAML), split per folder
postman-to-openapi source="Salesforce Platform APIs.postman_collection.json" output="api/openapi.yaml":
    #!/usr/bin/env bash
    set -euo pipefail
    tmp="$(mktemp --suffix=.json)"
    trap 'rm -f "$tmp"' EXIT
    node tools/postman-sanitize.js "{{source}}" "$tmp"
    mkdir -p "$(dirname "{{output}}")"
    npx -y -p postman-to-openapi@3.0.1 p2o "$tmp" -f "{{output}}" -o tools/p2o-options.json > /dev/null
    go run ./tools/openapi-split "{{output}}"

# Generate the chi server stubs and models from api/openapi.yaml
oapi-codegen:
    #!/usr/bin/env bash
    set -euo pipefail
    bin="$(mktemp -d)"
    trap 'rm -rf "$bin"' EXIT
    GOBIN="$bin" go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.2
    go run ./tools/openapi-codegen-split -oapi-codegen "$bin/oapi-codegen" -config api/oapi-codegen.yaml api/openapi.yaml

# Regenerate the OpenAPI spec from the Postman collection, then the server stubs
api: postman-to-openapi oapi-codegen
