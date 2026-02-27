/*
Package repositories provides a GORM-based base repository implementing generic CRUD, pagination, and raw SQL with context support.

Role in architecture:
  - Infrastructure adapter: implements persistence; can implement domain/port interfaces (e.g. GetByID) for use cases.

Responsibilities:
  - IBaseRepository: interface for Create, Save, Updates, Delete, First, Find, Scan, RawSQL, ExecSQL, IsEmpty, SimplePagination.
  - BaseRepository: implementation using getDB() (injected or database.GetDB/GetDBSite); all queries use WithContext(ctx).
  - Column name resolution from struct tags (cached by type); WHERE conditions support multi-placeholder keys.
  - SimplePagination: offset-based with optional preloads to avoid N+1; no business rules.

Constraints:
  - SQL purpose and query rules: First orders by created_at DESC and expects at most one row; SimplePagination applies LIMIT pageSize+1 and trims.
  - Error behavior: gorm.ErrRecordNotFound is translated to (notFound=true, err=nil) for First/Scan; other errors returned as-is.
  - No provider switching; DB is fixed at construction (global or injected).

This package must NOT:
  - Contain use-case or domain logic; only persistence mapping and query building.
*/
package repositories
