/*
Package mtls provides mutual TLS helpers for backend HTTPS servers and gateway outbound clients.

Role in architecture:
  - Infrastructure: builds tls.Config, configures http.Server / Gin listeners, and outbound transports.
  - Configuration is loaded via config.GetConfig().MTLS (or config.Setup).

Responsibilities:
  - ServerTLSConfig / ClientTLSConfig / CloneTLSConfig for certificate material.
  - ConfigureServer, ListenAndServe, RunGin (and *Config / *WithConfig variants) when MTLS_ENABLED.
  - NewTransport for outbound HTTP with client certificate authentication.

Constraints:
  - When MTLS_ENABLED is false, listen helpers use plain HTTP / no client cert attachment.
  - Pair server TLS with middlewares.MTLSMiddleware so application checks match client auth.
  - Skip paths for the middleware come from MTLS_SKIP_PATHS (e.g. /live, /ready, /metrics).

This package must NOT:
  - Contain use-case or domain logic.
  - Implement HTTP auth beyond TLS client certificate verification.
*/
package mtls
