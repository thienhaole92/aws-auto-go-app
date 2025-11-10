package config

import (
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/jackc/pgx/v5/tracelog"
)

type LogLevel tracelog.LogLevel

func (t *LogLevel) UnmarshalText(text []byte) error {
	l, err := tracelog.LogLevelFromString(strings.ToLower(string(text)))
	if err != nil {
		*t = LogLevel(tracelog.LogLevelError)
	}

	*t = LogLevel(l)

	return nil
}

type Config struct {
	// Common settings
	GracefulShutdownPeriod time.Duration `env:"GRACEFUL_SHUTDOWN_PERIOD" envDefault:"30s"`

	// Metric server settings
	MetricServerHost         string        `env:"METRIC_SERVER_HOST"          envDefault:"0.0.0.0"`
	MetricServerPort         int           `env:"METRIC_SERVER_PORT"          envDefault:"8082"`
	MetricServerReadTimeout  time.Duration `env:"METRIC_SERVER_READ_TIMEOUT"  envDefault:"30s"`
	MetricServerWriteTimeout time.Duration `env:"METRIC_SERVER_WRITE_TIMEOUT" envDefault:"30s"`

	// HTTP server settings
	HTTPServerHost         string        `env:"HTTP_SERVER_HOST"          envDefault:"0.0.0.0"`
	HTTPServerPort         int           `env:"HTTP_SERVER_PORT"          envDefault:"8081"`
	HTTPServerReadTimeout  time.Duration `env:"HTTP_SERVER_READ_TIMEOUT"  envDefault:"30s"`
	HTTPServerWriteTimeout time.Duration `env:"HTTP_SERVER_WRITE_TIMEOUT" envDefault:"30s"`
	HTTPEnableCORS         bool          `env:"HTTP_ENABLE_CORS"          envDefault:"false"`
	HTTPBodyLimit          string        `env:"HTTP_BODY_LIMIT"           envDefault:"100K"`
	HTTPSkipRequestID      bool          `env:"HTTP_SKIP_REQUEST_ID"      envDefault:"true"`

	// Keyval configuration
	KeyvalHost         string        `env:"KEYVAL_HOST"                 envDefault:"0.0.0.0"`
	KeyvalPort         int           `env:"KEYVAL_PORT"                 envDefault:"6379"`
	KeyvalPassword     string        `env:"KEYVAL_PASSWORD"`
	KeyvalDB           int           `env:"KEYVAL_DB"                   envDefault:"0"`
	KeyvalMaxIdleConns int           `env:"KEYVAL_MAX_IDLE_CONNS"       envDefault:"1"`
	KeyvalMinIdleConns int           `env:"KEYVAL_MIN_IDLE_CONNS"       envDefault:"1"`
	KeyvalPingTimeout  time.Duration `env:"KEYVAL_PING_TIMEOUT"         envDefault:"30s"`
	KeyvalDialTimeout  time.Duration `env:"KEYVAL_DIAL_TIMEOUT"         envDefault:"30s"`
	KeyvaReadTimeout   time.Duration `env:"KEYVAL_SERVER_READ_TIMEOUT"  envDefault:"30s"`
	KeyvaWriteTimeout  time.Duration `env:"KEYVAL_SERVER_WRITE_TIMEOUT" envDefault:"30s"`
	KeyvalPoolSize     int           `env:"KEYVAL_POOL_SIZE"            envDefault:"1"`

	// Postgres configuration
	PostgresURL                   string        `env:"POSTGRES_URL"                      envDefault:"0.0.0.0:5432"`
	PostgresMaxConnection         int32         `env:"POSTGRES_MAX_CONNECTION"           envDefault:"10"`
	PostgresMinConnection         int32         `env:"POSTGRES_MIN_CONNECTION"           envDefault:"1"`
	PostgresMaxConnectionIdleTime time.Duration `env:"POSTGRES_MAX_CONNECTION_IDLE_TIME" envDefault:"15m"`
	PostgresLogLevel              LogLevel      `env:"POSTGRES_LOG_LEVEL"                envDefault:"INFO"`

	// Migration settings
	MigrationEnabled bool   `env:"MIGRATION_ENABLED" envDefault:"false"`
	MigrationSource  string `env:"MIGRATION_SOURCE"  envDefault:"file://db/migrations"`
}

func New(opts *env.Options) (*Config, error) {
	cfg := new(Config)
	if opts != nil {
		if err := env.ParseWithOptions(cfg, *opts); err != nil {
			return nil, err
		}

		return cfg, nil
	}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
