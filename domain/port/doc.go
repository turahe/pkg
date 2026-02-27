/*
Package port defines repository and other ports (interfaces) used by use cases.
Implementations live in infrastructure (e.g. repositories package); domain does not depend on them.

Role in architecture:
  - Ports: interfaces that use cases depend on; implemented by adapters in other packages.

Responsibilities:
  - Define minimal, use-case-oriented interfaces (e.g. GetByID).
  - No methods except those required by use cases; avoid fat interfaces.

This package must NOT:
  - Import infrastructure packages or contain implementations.
  - Define business rules; only interface contracts.
*/
package port
