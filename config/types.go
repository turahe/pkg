package config

// Configuration holds all application configuration
type Configuration struct {
	Server       ServerConfiguration
	Cors         CorsConfiguration
	Database     DatabaseConfiguration
	DatabaseSite DatabaseConfiguration // Site database connection
	Redis        RedisConfiguration
	GCS          GCSConfiguration
	RateLimiter  RateLimiterConfiguration
	Timezone     TimezoneConfiguration
}

// ServerConfiguration holds server-related configuration
type ServerConfiguration struct {
	Port               string
	Secret             string
	Mode               string
	AccessTokenExpiry  int // hours
	RefreshTokenExpiry int // days
	SessionExpiry      int // hours, default 24
	SessionCookieName  string
	SessionSecure      bool   // Set Secure flag for cookies (HTTPS only)
	SessionHttpOnly    bool   // Set HttpOnly flag for cookies (default true)
	SessionSameSite    string // SameSite cookie attribute: "strict", "lax", "none"
}

// CorsConfiguration holds CORS-related configuration
type CorsConfiguration struct {
	Global bool
	Ips    string
}

// DatabaseConfiguration holds database-related configuration
type DatabaseConfiguration struct {
	Driver                string
	Dbname                string
	Username              string
	Password              string
	Host                  string
	Port                  string
	Sslmode               bool
	Logmode               bool
	CloudSQLInstance      string `mapstructure:"cloud_sql_instance"` // Cloud SQL instance connection name (format: project:region:instance)
	MaxIdleConns           int // Max idle connections in pool (0 = default 5)
	MaxOpenConns           int // Max open connections (0 = default 10)
	ConnMaxLifetimeMinutes int // Max lifetime of a connection in minutes (0 = default 1440 = 24h)
}

// RedisConfiguration holds Redis-related configuration
type RedisConfiguration struct {
	Enabled      bool
	Host         string
	Port         string
	Password     string
	DB           int
	ClusterMode  bool   // Enable cluster mode for Google Cloud Memorystore Redis Cluster
	ClusterNodes string // Comma-separated list of cluster node addresses (e.g., "10.0.0.1:6379,10.0.0.2:6379")
}

// GCSConfiguration holds Google Cloud Storage-related configuration
type GCSConfiguration struct {
	Enabled         bool
	BucketName      string
	CredentialsFile string // Path to service account JSON file (optional, uses ADC if not provided)
}

// RateLimiterConfiguration holds rate limiter-related configuration
type RateLimiterConfiguration struct {
	Enabled   bool
	Requests  int    // Number of requests allowed per window
	Window    int    // Time window in seconds
	KeyBy     string // Key strategy: "ip" (default) or "user" (requires auth)
	SkipPaths string // Comma-separated list of paths to skip rate limiting (e.g., "/health,/metrics")
}

// TimezoneConfiguration holds timezone-related configuration
type TimezoneConfiguration struct {
	Timezone string // IANA timezone identifier (e.g., "Asia/Jakarta", "UTC", "America/New_York")
}
