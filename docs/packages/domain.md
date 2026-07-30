# domain

Shared domain sentinel errors with **no infrastructure dependencies**.

**Import:** `github.com/turahe/pkg/domain`

## Exports

```go
var (
    ErrNotFound     = errors.New("not found")
    ErrUnauthorized = errors.New("unauthorized")
)
```

Handlers map these via `handler.BaseHandler.HandleServiceError` (also re-exported on `handler`).

## Constraints

- Must not import database, redis, config, or HTTP packages.
- No business logic beyond shared error values — domain entities live in consuming apps.

---

## domain/port

**Import:** `github.com/turahe/pkg/domain/port`

Use-case **ports** (interfaces). Implementations live in repositories or adapters in your application.

```go
type GetByID interface {
    GetByID(ctx context.Context, id string) (interface{}, bool, error)
}
```

**Contract:** never return `(nil, true, err)`. Prefer `(value, true, nil)`, `(nil, false, nil)` for not found, or `(nil, false, err)` on failure.

### Constraints

- Must not import infrastructure.
- Must not contain implementations or business rules — interfaces only.

## See also

- [usecase](usecase.md)
- [repositories](repositories.md)
- [handler](handler.md)
- [Architecture](../architecture.md)
