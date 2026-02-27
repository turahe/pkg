/*
Package response provides standardized JSON response types and helpers using a 7-digit composite code (HTTP_STATUS + SERVICE_CODE + CASE_CODE).

Role in architecture:
  - Shared adapter output: used by handlers to write consistent JSON and by handler package for error mapping.

Responsibilities:
  - Build and parse composite response codes (codes.go).
  - Write success, error, validation, pagination, and custom responses to Gin context (response.go, validation.go, pagination.go).
  - Format go-playground/validator errors into Laravel-style field maps (validation.go).

Constraints:
  - No business logic; only response shape and code building.
  - Service and case codes are defined in this package; callers pass them from response package constants.

This package must NOT:
  - Perform I/O except writing to gin.Context.
  - Depend on domain or use-case packages.
*/
package response
