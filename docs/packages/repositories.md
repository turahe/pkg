# repositories

Generic GORM CRUD and pagination base repository.

**Import:** `github.com/turahe/pkg/repositories`

## Role

Persistence adapter. Implement app-specific repos by embedding `BaseRepository` or composing with an injected `*gorm.DB`. No use-case/domain logic here.

## Constructors

```go
repo := repositories.NewBaseRepository()              // primary global DB
site := repositories.NewSiteBaseRepository()          // site global DB
repo := repositories.NewBaseRepositoryWithDB(db)      // DI (preferred)
site := repositories.NewSiteBaseRepositoryWithDB(db)
```

## Interface (`IBaseRepository`)

| Method | Notes |
|--------|-------|
| `Create`, `Save`, `Updates`, `Delete` | Standard GORM write paths |
| `First` | Orders `created_at DESC`; `ErrRecordNotFound` → `(notFound=true, nil)` |
| `Find`, `Scan` | Lists / custom scan |
| `RawSQL`, `ExecSQL` | Escapes for advanced queries |
| `IsEmpty` | |
| `SimplePagination` | `LIMIT pageSize+1` for has-more detection; supports preloads |

Use `types.Conditions` for where maps:

```go
repo.First(ctx, &out, types.Conditions{"id = ?": id})
repo.SimplePagination(ctx, &model, &out, page, size, conditions, orders, "User", "Items")
```

## Constraints

- **Condition keys are SQL.** `types.Conditions` keys (`"email = ?"`) are interpolated; only values are bound. Keys must be developer-written constants, never request input.
- **Orders are validated.** `Find`, `Scan` and `SimplePagination`/`Paginate` return `ErrInvalidOrder` unless each order is a column list like `"name ASC, users.created_at DESC"` (optional `NULLS FIRST|LAST`). Still map user-facing sort params through an allowlist of columns; use `repositories.ValidateOrder` to check early.
- Prefer injected DB over globals in new code.
- Map not-found to domain errors in the use-case or handler layer as needed.
- Must not import use-case or contain business rules.

## See also

- [database](database.md)
- [types](types.md)
- [domain/port](domain.md#domainport)
