/*
Package config provides application configuration loaded from environment variables.

Role in architecture:
  - Infrastructure: reads from OS environment and optional .env file (via godotenv).
  - Single source of truth for server, database, Redis, GCS, rate limiter, CORS, and timezone settings.

Responsibilities:
  - Load and parse environment variables into typed structs (Configuration and nested types).
  - Validate required database configuration (Dbname, Username, Password; CloudSQLInstance when driver is cloudsql-*).
  - Expose global configuration via GetConfig(); lazy-build from env when Config is nil.
  - Strip surrounding quotes from env values; treat known placeholders as invalid.

Constraints:
  - No file-based config formats (YAML/JSON) except .env for variable loading.
  - No secret injection from external secret managers; callers must set env before Setup/GetConfig.
  - Database validation runs only for primary and (if Dbname set) site database; other sections are not validated.

This package must NOT:
  - Perform I/O beyond reading environment and loading a single .env file.
  - Depend on database, Redis, or HTTP packages.
*/
package config
