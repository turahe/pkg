package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Placeholder values that should be treated as invalid
var invalidPlaceholders = map[string]bool{
	"your_database_name":     true,
	"your_database_user":     true,
	"your_database_password": true,
}

// buildConfigFromEnv builds configuration directly from environment variables
func buildConfigFromEnv() *Configuration {
	return &Configuration{
		Server: ServerConfiguration{
			Port:                getEnvOrDefault("SERVER_PORT", "8080"),
			Mode:                getEnvOrDefault("SERVER_MODE", "debug"),
			AccessTokenExpiry:   parseInt("SERVER_ACCESS_TOKEN_EXPIRY", 1),
			RefreshTokenExpiry:  parseInt("SERVER_REFRESH_TOKEN_EXPIRY", 7),
			SessionExpiry:       parseInt("SERVER_SESSION_EXPIRY", 24),
			SessionCookieName:   getEnvOrDefault("SERVER_SESSION_COOKIE_NAME", "admin_session"),
			SessionSecure:       parseBool("SERVER_SESSION_SECURE", false),
			SessionHttpOnly:     parseBool("SERVER_SESSION_HTTP_ONLY", true),
			SessionSameSite:     getEnvOrDefault("SERVER_SESSION_SAME_SITE", "lax"),
			JWTSigningAlgorithm: getEnvOrDefault("JWT_SIGNING_ALGORITHM", "RS256"),
			JWTPrivateKey:       getEnvOrDefault("JWT_PRIVATE_KEY", ""),
			JWTPublicKey:        getEnvOrDefault("JWT_PUBLIC_KEY", ""),
			JWTIssuer:           getEnvOrDefault("JWT_ISSUER", ""),
			JWTAudience:         getEnvOrDefault("JWT_AUDIENCE", ""),
			JWTKeyID:            getEnvOrDefault("JWT_KEY_ID", ""),
		},
		Cors: CorsConfiguration{
			Global:   parseBool("CORS_GLOBAL", true),
			Frontend: getEnvOrDefault("CORS_FRONTEND", ""),
			Ips:      getEnvOrDefault("CORS_IPS", ""),
		},
		Database: DatabaseConfiguration{
			Driver:                 getEnvOrDefault("DATABASE_DRIVER", "mysql"),
			Dbname:                 getEnvOrDefault("DATABASE_DBNAME", ""),
			Username:               getEnvOrDefault("DATABASE_USERNAME", ""),
			Password:               getEnvOrDefault("DATABASE_PASSWORD", ""),
			Host:                   getEnvOrDefault("DATABASE_HOST", "127.0.0.1"),
			Port:                   getEnvOrDefault("DATABASE_PORT", "3306"),
			Sslmode:                parseBool("DATABASE_SSLMODE", false),
			Logmode:                parseBool("DATABASE_LOGMODE", true),
			CloudSQLInstance:       getEnvOrDefault("DATABASE_CLOUD_SQL_INSTANCE", ""),
			MaxIdleConns:           parseInt("DATABASE_MAX_IDLE_CONNS", 5),
			MaxOpenConns:           parseInt("DATABASE_MAX_OPEN_CONNS", 10),
			ConnMaxLifetimeMinutes: parseInt("DATABASE_CONN_MAX_LIFETIME", 1440),
			ConnectionTimezone:     getEnvOrDefault("DATABASE_TIMEZONE", ""),
		},
		DatabaseSite: DatabaseConfiguration{
			Driver:                 getEnvOrDefault("DATABASE_DRIVER_SITE", "mysql"),
			Dbname:                 getEnvOrDefault("DATABASE_DBNAME_SITE", ""),
			Username:               getEnvOrDefault("DATABASE_USERNAME_SITE", ""),
			Password:               getEnvOrDefault("DATABASE_PASSWORD_SITE", ""),
			Host:                   getEnvOrDefault("DATABASE_HOST_SITE", "127.0.0.1"),
			Port:                   getEnvOrDefault("DATABASE_PORT_SITE", "3306"),
			Sslmode:                parseBool("DATABASE_SSLMODE_SITE", false),
			Logmode:                parseBool("DATABASE_LOGMODE_SITE", true),
			CloudSQLInstance:       getEnvOrDefault("DATABASE_CLOUD_SQL_INSTANCE_SITE", ""),
			MaxIdleConns:           parseInt("DATABASE_MAX_IDLE_CONNS_SITE", 5),
			MaxOpenConns:           parseInt("DATABASE_MAX_OPEN_CONNS_SITE", 10),
			ConnMaxLifetimeMinutes: parseInt("DATABASE_CONN_MAX_LIFETIME_SITE", 1440),
			ConnectionTimezone:     getEnvOrDefault("DATABASE_TIMEZONE_SITE", ""),
		},
		Redis: RedisConfiguration{
			Enabled:         parseBool("REDIS_ENABLED", false),
			Host:            getEnvOrDefault("REDIS_HOST", "127.0.0.1"),
			Port:            getEnvOrDefault("REDIS_PORT", "6379"),
			Password:        getEnvOrDefault("REDIS_PASSWORD", ""),
			DB:              parseInt("REDIS_DB", 1),
			ClusterMode:     parseBool("REDIS_CLUSTER_MODE", false),
			ClusterNodes:    getEnvOrDefault("REDIS_CLUSTER_NODES", ""),
			PoolSize:        parseInt("REDIS_POOL_SIZE", 0),
			MinIdleConns:    parseInt("REDIS_MIN_IDLE_CONNS", 0),
			ReadTimeoutSec:  parseInt("REDIS_READ_TIMEOUT_SEC", 0),
			WriteTimeoutSec: parseInt("REDIS_WRITE_TIMEOUT_SEC", 0),
		},
		GCS: GCSConfiguration{
			Enabled:         parseBool("GCS_ENABLED", false),
			BucketName:      getEnvOrDefault("GCS_BUCKET_NAME", ""),
			CredentialsFile: getEnvOrDefault("GCS_CREDENTIALS_FILE", ""),
		},
		RateLimiter: RateLimiterConfiguration{
			Enabled:   parseBool("RATE_LIMITER_ENABLED", false),
			Requests:  parseInt("RATE_LIMITER_REQUESTS", 100),
			Window:    parseInt("RATE_LIMITER_WINDOW", 60),
			KeyBy:     getEnvOrDefault("RATE_LIMITER_KEY_BY", "ip"),
			SkipPaths: getEnvOrDefault("RATE_LIMITER_SKIP_PATHS", ""),
		},
		Timezone: TimezoneConfiguration{
			Timezone: getEnvOrDefault("SERVER_TIMEZONE", "UTC"),
		},
		OpenTelemetry: OpenTelemetryConfiguration{
			Exporter:         getEnvOrDefault("OTEL_TRACES_EXPORTER", "otlp"),
			Endpoint:         getEnvOrDefault("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
			TracesEndpoint:   getEnvOrDefault("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", ""),
			Insecure:         parseBool("OTEL_EXPORTER_OTLP_INSECURE", true),
			Headers:          parseKeyValueHeaders(getEnvOrDefault("OTEL_EXPORTER_OTLP_HEADERS", "")),
			ServiceName:      getEnvOrDefault("OTEL_SERVICE_NAME", ""),
			Environment:      firstNonEmpty(getEnvOrDefault("OTEL_ENVIRONMENT", ""), getEnvOrDefault("APP_ENV", "")),
			ServiceVersion:   firstNonEmpty(getEnvOrDefault("OTEL_SERVICE_VERSION", ""), getEnvOrDefault("SENTRY_RELEASE", "")),
			TracesSamplerArg: parseFloatRatio("OTEL_TRACES_SAMPLER_ARG", 1.0),
			ShutdownTimeout:  parseDuration("OTEL_SHUTDOWN_TIMEOUT", 5*time.Second),
			GORMEnabled:      parseBool("OTEL_GORM_ENABLED", true),
			GCPProjectID:     firstNonEmpty(getEnvOrDefault("OTEL_GCP_PROJECT_ID", ""), getEnvOrDefault("GOOGLE_CLOUD_PROJECT", "")),
			GCPPropagator:    parseBool("OTEL_GCP_PROPAGATOR", false),
		},
		Sentry: SentryConfiguration{
			DSN:              getEnvOrDefault("SENTRY_DSN", ""),
			Environment:      firstNonEmpty(getEnvOrDefault("SENTRY_ENVIRONMENT", ""), getEnvOrDefault("APP_ENV", "")),
			Release:          getEnvOrDefault("SENTRY_RELEASE", ""),
			ServerName:       getEnvOrDefault("SENTRY_SERVER_NAME", ""),
			Debug:            parseBool("SENTRY_DEBUG", false),
			AttachStacktrace: parseBool("SENTRY_ATTACH_STACKTRACE", true),
			SampleRate:       parseFloatRatio("SENTRY_SAMPLE_RATE", 1.0),
			TracesSampleRate: parseFloatRatio("SENTRY_TRACES_SAMPLE_RATE", 0.0),
			FlushTimeout:     parseDuration("SENTRY_FLUSH_TIMEOUT", 2*time.Second),
		},
		MTLS: MTLSConfiguration{
			Enabled:        parseBool("MTLS_ENABLED", false),
			CACertFile:     getEnvOrDefault("MTLS_CA_CERT", "/etc/mtls/ca.crt"),
			ServerCertFile: getEnvOrDefault("MTLS_SERVER_CERT", "/etc/mtls/server.crt"),
			ServerKeyFile:  getEnvOrDefault("MTLS_SERVER_KEY", "/etc/mtls/server.key"),
			ClientCertFile: getEnvOrDefault("MTLS_CLIENT_CERT", "/etc/mtls/gateway.crt"),
			ClientKeyFile:  getEnvOrDefault("MTLS_CLIENT_KEY", "/etc/mtls/gateway.key"),
			SkipPaths:      getEnvOrDefault("MTLS_SKIP_PATHS", "/live,/ready,/metrics"),
		},
	}
}

// getEnvOrDefault gets environment variable or returns default value.
// Also strips surrounding quotes if present.
func getEnvOrDefault(key, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}

	// Strip surrounding quotes if present
	if len(val) >= 2 {
		if (val[0] == '"' && val[len(val)-1] == '"') ||
			(val[0] == '\'' && val[len(val)-1] == '\'') {
			return val[1 : len(val)-1]
		}
	}
	return val
}

// parseBool parses a boolean from environment variable
func parseBool(key string, defaultValue bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	result, err := strconv.ParseBool(val)
	if err != nil {
		return defaultValue
	}
	return result
}

// parseInt parses an integer from environment variable
func parseInt(key string, defaultValue int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	result, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func parseFloatRatio(key string, defaultValue float64) float64 {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return defaultValue
	}
	result, err := strconv.ParseFloat(val, 64)
	if err != nil || result < 0 || result > 1 {
		return defaultValue
	}
	return result
}

func parseDuration(key string, defaultValue time.Duration) time.Duration {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return defaultValue
	}
	result, err := time.ParseDuration(val)
	if err != nil || result <= 0 {
		return defaultValue
	}
	return result
}

func parseKeyValueHeaders(raw string) map[string]string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	headers := make(map[string]string)
	for _, pair := range strings.Split(raw, ",") {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		k, v, ok := strings.Cut(pair, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if k != "" {
			headers[k] = v
		}
	}
	if len(headers) == 0 {
		return nil
	}
	return headers
}
