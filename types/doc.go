/*
Package types provides shared types used across handler, repository, and use-case layers.

Role in architecture:
  - Shared kernel: no dependencies on infrastructure; used by domain-facing code and adapters.

Responsibilities:
  - Conditions: map type for repository WHERE clauses (key = SQL fragment, value = arg).
  - TimeRange: start/end string pair for time-bounded queries.
  - PageInfo: pagination request (pageNumber, pageSize) with form/json tags.

This package must NOT:
  - Contain business rules or validation logic.
*/
package types
