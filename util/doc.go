/*
Package util provides generic helpers for common operations: empty checks, slice membership, deduplication, and phone formatting.

Role in architecture:
  - Shared utilities: no business logic; pure functions used by handlers, use cases, or other packages.

Responsibilities:
  - IsEmpty: reflect-based check for zero/nil/empty values (string, numeric, bool, ptr, slice, map).
  - InAnySlice, RemoveDuplicates: generic slice helpers.
  - FormatPhoneNumber: E.164 formatting via nyaruka/phonenumbers with optional default region.

This package must NOT:
  - Depend on database, HTTP, or config; only standard library and phonenumbers.
*/
package util
