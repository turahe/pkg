/*
Package usecase provides the application service layer: use cases depend only on domain (errors and port interfaces).

Role in architecture:
  - Application core: handlers call use cases with context; use cases call repository ports and return domain errors.

Responsibilities:
  - Runner: interface for use cases that run with context and return an error (e.g. for handler to map via HandleServiceError).
  - Func: adapter from a function to Runner.
  - GetItemByID: example use case that calls port.GetByID and returns domain.ErrNotFound when not found.

Constraints:
  - Use cases must not import handler, database, redis, or config; only domain and domain/port.
  - No business logic in this package beyond example; real use cases live in the application.

This package must NOT:
  - Perform I/O directly; only through port interfaces.
  - Depend on infrastructure packages.
*/
package usecase
