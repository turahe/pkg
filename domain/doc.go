/*
Package domain provides domain errors and port interfaces. It has no dependencies on
infrastructure (database, HTTP, config), so use cases can depend only on domain.

Role in architecture:
  - Core: defines shared errors and interfaces that use cases depend on.
  - Implementations of ports live in infrastructure (e.g. repositories package).

Responsibilities:
  - Define sentinel errors (ErrNotFound, ErrUnauthorized) for use-case and handler error handling.
  - Define port interfaces (e.g. GetByID in subpackage port) that use cases call and infrastructure implements.

This package must NOT:
  - Import database, redis, config, or HTTP packages.
  - Contain business logic; it only defines contracts and errors.
*/
package domain
