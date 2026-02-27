package config

// Configuration holds the full application configuration. All sections are
// populated from environment variables by buildConfigFromEnv (see parser.go).
type Configuration struct {
	Server       ServerConfiguration
	Cors         CorsConfiguration
	Database     DatabaseConfiguration
	DatabaseSite DatabaseConfiguration // Optional second database; leave Dbname empty to disable.
	Redis        RedisConfiguration
	GCS          GCSConfiguration
	RateLimiter  RateLimiterConfiguration
	Timezone     TimezoneConfiguration
}

// ServerConfiguration holds server and session settings (port, secret, mode, token and session expiry).
type ServerConfiguration struct {
	Port               string // Listen port; default "8080"
	Secret             string // JWT signing secret; required for auth
	Mode               string // Gin mode: "debug", "release", "test"
	AccessTokenExpiry  int    // Access token lifetime in hours
	RefreshTokenExpiry int    // Refresh token lifetime in days
	SessionExpiry      int    // Session lifetime in hours; default 24
	SessionCookieName  string
	SessionSecure      bool   // Secure flag for cookies (HTTPS only)
	SessionHttpOnly    bool   // HttpOnly flag; default true
	SessionSameSite    string // "strict", "lax", or "none"
}

// CorsConfiguration holds CORS settings. Global true allows all origins; Ips is used when Global is false.
type CorsConfiguration struct {
	Global bool   // If true, allow all origins
	Ips    string // Comma-separated allowed IPs when Global is false
}

// DatabaseConfiguration holds database connection and pool settings. Required: Dbname, Username, Password.
// For Cloud SQL drivers (cloudsql-mysql, cloudsql-postgres), CloudSQLInstance (project:region:instance) is required.
type DatabaseConfiguration struct {
	Driver                string
	Dbname                string
	Username              string
	Password              string
	Host                  string
	Port                  string
	Sslmode               bool
	Logmode               bool
	CloudSQLInstance      string `mapstructure:"cloud_sql_instance"` // project:region:instance for Cloud SQL
	MaxIdleConns          int    // 0 = default 5
	MaxOpenConns          int    // 0 = default 10
	ConnMaxLifetimeMinutes int   // 0 = default 1440 (24h)
}

// RedisConfiguration holds Redis connection and pool settings. Set Enabled true to use Redis.
type RedisConfiguration struct {
	Enabled          bool
	Host             string
	Port             string
	Password         string
	DB               int    // Database index; ignored in cluster mode
	ClusterMode      bool   // Use cluster client (e.g. Google Cloud Memorystore)
	ClusterNodes     string // Comma-separated host:port when ClusterMode is true
	PoolSize         int    // 0 = client default; use e.g. 100 for high RPS
	MinIdleConns     int    // 0 = default
	ReadTimeoutSec   int    // 0 = no timeout
	WriteTimeoutSec  int    // 0 = no timeout
}

// GCSConfiguration holds Google Cloud Storage settings. BucketName required when Enabled is true.
type GCSConfiguration struct {
	Enabled         bool
	BucketName      string
	CredentialsFile string // Optional; omit to use Application Default Credentials
}

// RateLimiterConfiguration holds rate limiter settings. Requires Redis when Enabled is true.
type RateLimiterConfiguration struct {
	Enabled   bool
	Requests  int    // Max requests per window
	Window    int    // Window size in seconds
	KeyBy     string // "ip" or "user" (user requires auth middleware)
	SkipPaths string // Comma-separated paths to skip (e.g. "/health,/metrics")
}

// TimezoneConfiguration holds the server timezone (IANA name, e.g. "Asia/Jakarta", "UTC").
type TimezoneConfiguration struct {
	Timezone string
}
