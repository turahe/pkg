/*
Package handler provides a base HTTP handler for Gin with binding, validation, error mapping, and pagination helpers.

Role in architecture:
  - Adapter: translates HTTP (Gin context) into use-case inputs and maps use-case/domain errors to HTTP responses.

Responsibilities:
  - Bind request body, query, and URI parameters by content-type and method.
  - Map domain errors (ErrNotFound, ErrUnauthorized) and optional notFoundMessages to standardized JSON responses (response package).
  - Normalize pagination parameters and build pagination responses.
  - Extract user ID and roles from context set by auth middleware.

This package must NOT:
  - Contain business logic; only request/response mapping and validation response formatting.
  - Depend on concrete repositories or use cases; it receives errors and data from callers.
*/
package handler
