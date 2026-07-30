# types

Shared helper types for repository, handler, and use-case layers.

**Import:** `github.com/turahe/pkg/types`

## Exports

```go
type Conditions map[string]interface{}  // GORM-style where fragments → args

type PageInfo struct {
    PageNumber int
    PageSize   int
}

type TimeRange struct {
    Start time.Time
    End   time.Time
}
```

Example:

```go
repo.First(ctx, &out, types.Conditions{"id = ?": id, "status = ?": "active"})
```

## Constraints

- No business rules or validation logic — plain data carriers only.

## See also

- [repositories](repositories.md)
- [handler](handler.md)
