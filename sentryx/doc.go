/*
Package sentryx wraps github.com/getsentry/sentry-go with configuration from config
so the rest of the service does not need to know about Sentry's surface area.

Role in architecture:
  - Infrastructure: initializes the global Sentry hub and provides Flush for shutdown.
  - Settings are loaded via config.GetConfig().Sentry (or config.Setup).

Responsibilities:
  - Init: configure DSN, environment, release, server name, sample rates, and stack traces.
  - LoadConfig: read SentryConfiguration from the global config package.
  - Flush: drain buffered events with a timeout (safe when Sentry was never initialized).

Constraints:
  - SENTRY_DSN empty is the explicit "Sentry is off" signal — Init is a no-op and returns false.
  - Init returns bool (enabled), not error; invalid DSN is logged and treated as off.
  - The service must keep running whether Sentry is on, off, or misconfigured.
  - Default server name is "app" when SENTRY_SERVER_NAME is empty.
  - Sample rates outside [0, 1] are clamped to safe defaults (1.0 errors, 0.0 traces).

This package must NOT:
  - Contain use-case or domain logic.
  - Fail process startup when Sentry cannot be enabled.
*/
package sentryx
