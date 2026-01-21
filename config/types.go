package config

// Configuration holds all application configuration
type Configuration struct {
	Server       ServerConfiguration
	Cors         CorsConfiguration
	Database     DatabaseConfiguration
	DatabaseSite DatabaseConfiguration // Site database connection
	Redis        RedisConfiguration
	GCS          GCSConfiguration
}

// ServerConfiguration holds server-related configuration
type ServerConfiguration struct {
	Port               string
	Secret             string
	Mode               string
	AccessTokenExpiry  int // hours
	RefreshTokenExpiry int // days
}

// CorsConfiguration holds CORS-related configuration
type CorsConfiguration struct {
	Global bool
	Ips    string
}

// DatabaseConfiguration holds database-related configuration
type DatabaseConfiguration struct {
	Driver           string
	Dbname           string
	Username         string
	Password         string
	Host             string
	Port             string
	Sslmode          bool
	Logmode          bool
	CloudSQLInstance string `mapstructure:"cloud_sql_instance"` // Cloud SQL instance connection name (format: project:region:instance)
}

// RedisConfiguration holds Redis-related configuration
type RedisConfiguration struct {
	Enabled  bool
	Host     string
	Port     string
	Password string
	DB       int
}

// GCSConfiguration holds Google Cloud Storage-related configuration
type GCSConfiguration struct {
	Enabled         bool
	BucketName      string
	CredentialsFile string // Path to service account JSON file (optional, uses ADC if not provided)
}
