/*
Package jwt provides JWT token generation and validation (HS256) and bcrypt password comparison using config for secret and expiry.

Role in architecture:
  - Infrastructure: used by auth middleware and handlers; reads config.Server.Secret and expiry settings.

Responsibilities:
  - Init: load secret and expiry from config; panic if missing.
  - GenerateToken, GenerateTokenWithExpiry, GenerateRefreshToken: issue signed tokens with Claims.UUID.
  - ValidateToken: parse and verify; return Claims or error.
  - ComparePassword: bcrypt comparison.
  - GetCurrentUserUUID: read user_id from Gin context (string or uuid.UUID) and return uuid.UUID.

Constraints:
  - Single secret and expiry from config; no key rotation or multi-tenant secrets in this package.
  - Signing method is HS256 only.

This package must NOT:
  - Contain use-case logic; only token and password operations.
*/
package jwt
