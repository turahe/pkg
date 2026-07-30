# usecase

Application-layer scaffolding. Depends on `domain` / `domain/port` only — never on Gin, GORM, Redis, or config.

**Import:** `github.com/turahe/pkg/usecase`

## Role

Orchestrate ports. Real business logic belongs in your application's use-case packages; this module provides patterns and a small example.

## Types

```go
type Runner interface {
    Run(ctx context.Context) error
}

type Func func(ctx context.Context) error

func (f Func) Run(ctx context.Context) error { return f(ctx) }
```

## Example use case

```go
item, err := usecase.GetItemByID(ctx, repo, id)
// repo is port.GetByID
// maps not-found bool to domain.ErrNotFound
```

## Constraints

- Must not import `handler`, `database`, `redis`, or `config`.
- I/O only through ports.
- Keep side effects and transactions at the edges (adapters).

## See also

- [domain/port](domain.md#domainport)
- [Architecture](../architecture.md)
