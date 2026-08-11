# Development

## Makefile targets

```bash
make help            # list targets
make build           # go build ./...
make test            # unit/integration (local services if needed)
make test-race       # CGO_ENABLED=1 race detector
make test-cover      # coverage summary → coverage.out
make test-docker     # full suite in Docker (Redis + MySQL + Postgres)
make services-up     # docker compose up -d (local Redis/MySQL/Postgres)
make services-down   # stop local services
make lint            # golangci-lint via Docker (v2.12.2, matches CI)
make vuln            # govulncheck
make tidy            # go mod tidy && verify
make docker-build    # production image
make clean           # remove coverage.out / test.log
```

## Testing

```bash
go test ./...
CGO_ENABLED=1 go test -race ./...
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

Integration tests for Redis / MySQL / Postgres **skip** when services are unreachable. CI and `make test-docker` run the full matrix.

Packages with tests: `config`, `crypto`, `database`, `gcs`, `handler`, `jwt`, `logger`, `middlewares`, `mtls`, `otelx`, `redis`, `repositories`, `response`, `sentryx`, `types`, `util`.

## Local services

```bash
docker compose up -d   # or: make services-up
```

See [`docker-compose.yml`](../docker-compose.yml) and [`docker-compose.test.yml`](../docker-compose.test.yml).

## Lint & security

Config: [`.golangci.yml`](../.golangci.yml).

```bash
make lint
make vuln
```

## Documentation conventions

- Every package has a **`doc.go`** (role, responsibilities, constraints, must-not).
- Exported symbols start comments with the symbol name (GoDoc style).
- Human guides live under [`docs/`](README.md); keep them aligned with code and `.env.example`.
- Prefer DI APIs (`database.New`, `jwt.NewManager`) in new examples over legacy globals.

## Viewing GoDoc locally

```bash
go doc github.com/turahe/pkg/database
# or
go install golang.org/x/pkgsite/cmd/pkgsite@latest
pkgsite -http=:8080
```

Published docs: [pkg.go.dev/github.com/turahe/pkg](https://pkg.go.dev/github.com/turahe/pkg).
