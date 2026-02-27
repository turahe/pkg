/*
Package crypto provides password hashing and comparison using bcrypt (golang.org/x/crypto/bcrypt).

Role in architecture:
  - Infrastructure utility: no business logic; used by auth or user flows that need to hash/compare passwords.

Responsibilities:
  - HashAndSalt: hash a plain password with bcrypt.MinCost; log and return empty string on error.
  - ComparePassword: return true if plain password matches the hash; log on error.

Constraints:
  - Single cost (MinCost); no configurable cost or salt in this package.
  - Depends on logger for error logging.

This package must NOT:
  - Store or retrieve passwords; only hash and compare.
*/
package crypto
