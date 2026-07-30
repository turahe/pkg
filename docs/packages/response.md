# response

Standardized Gin JSON responses with a **7-digit composite code**: `HTTP_STATUS` (3) + `SERVICE_CODE` (2) + `CASE_CODE` (2).

**Import:** `github.com/turahe/pkg/response`

## Role

Shared adapter output shape. No business logic; no I/O except writing `gin.Context`.

## Composite codes

```go
code := response.BuildResponseCode(201, response.ServiceCodeUser, response.CaseCodeCreated)
// e.g. 2010301
httpStatus, service, caseCode := response.ParseResponseCode(code)
```

Service codes (`ServiceCode*`) and case codes (`CaseCode*`) are package constants (00–21 services, 01–99 cases — see `codes.go`).

## Writers

| Helper | Typical use |
|--------|-------------|
| `Ok`, `OkWithCode`, `Result`, `ResultWithCode` | 200 success |
| `Created`, `Updated`, `Deleted` | Mutations |
| `Fail`, `FailWithCode` | Errors |
| `ValidationError`, `ValidationError*` | 422 + field map |
| `UnauthorizedError`, `NotFoundError`, `ConflictError`, `ForbiddenError` | Common HTTP errors |
| `CursorPaginated`, `SimplePaginated` | List endpoints |

Validation formatting: `FormatValidationError`, `GetJSONFieldName` (Laravel-style field maps from go-playground/validator).

Pagination defaults: `DefaultPageNumber`, `DefaultPageSize`, `MaxPageSize`.

Types: `CommonResponse`, `ValidationErrorResponse`, `CursorPaginationResponse`, `SimplePaginationResponse`, and paginated wrappers.

## Constraints

- Callers pass service/case codes from this package's constants.
- Must not depend on domain or use-case packages.

## See also

- [handler](handler.md)
