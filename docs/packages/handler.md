# handler

Base Gin handler helpers: bind/validate, map domain errors, pagination, identity from context.

**Import:** `github.com/turahe/pkg/handler`

## Role

HTTP adapter utilities. Embed `BaseHandler` in your handlers; do not put business logic here.

## BaseHandler

```go
type MyHandler struct {
    handler.BaseHandler
    // deps...
}

func (h *MyHandler) Get(c *gin.Context) {
    id, err := h.GetIDFromParam(c, "id")
    if err != nil {
        return
    }
    // call use case...
    if err != nil {
        h.HandleServiceError(c, err)
        return
    }
    response.Ok(c, data)
}
```

| Method | Purpose |
|--------|---------|
| `GetTraceID` | Trace ID from context |
| `ValidateReqParams` | Bind + validate |
| `HandleValidationError` | Write validation response |
| `NormalizePagination` | Clamp page/size |
| `GetIDFromParam` / `GetIDFromRequestOrParam` | Path / body ID |
| `HandleServiceError` | Map `domain.ErrNotFound` / `ErrUnauthorized` etc. |
| `BuildPaginationResponse` | Pagination payload helper |
| `GetCurrentUserID` / `CheckUserHasRole` | Auth context |

Re-exports: `handler.ErrNotFound`, `handler.ErrUnauthorized` (domain sentinels).

## Constraints

- Must not depend on concrete repositories or use cases (inject interfaces).
- Must not contain business rules.

## See also

- [response](response.md)
- [domain](domain.md)
- [usecase](usecase.md)
