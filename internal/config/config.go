package config

import (
	"clearway-test-task/pkg/validator"
	"fmt"
	"time"
)

type Http struct {
	// NET_HOST. Address to listen. Default to 0.0.0.0
	Host string `mapstructure:"net_host" validate:"ip4_addr|hostname_rfc1123"`
	// NET_PORT. Port to listen. Privileged ports aren't allowed. Default to 8080
	Port int `mapstructure:"net_port" validate:"numeric,gt=1024,lt=65536" env:"NET_PORT"`
	// NET_READ_TIMEOUT. Read connection timeout. Min 100 ms. Default to 500 ms
	ReadTimeout time.Duration `mapstructure:"net_read_timeout" validate:"min=500ms,max=1m"`
	// NET_WRITE_TIMEOUT. Write connection timeout. Min 100 ms. Default to 500 ms
	WriteTimeout time.Duration `mapstructure:"net_write_timeout" validate:"min=500ms,max=1m"`
	// NET_IDLE_TIMEOUT. Idle connection timeout. Min 100 ms. Default to 1 s
	IdleTimeout time.Duration `mapstructure:"net_idle_timeout" validate:"min=1s,max=1m"`
}

type Logging struct {
	// LOG_FORMAT. Log format. Default to text
	Format string `mapstructure:"log_format" validate:"oneof=text json"`
	// LOG_LEVEL. Log level. Default to info
	Level string `mapstructure:"log_level" validate:"oneof=debug info warn error"`
}

type Auth struct {
	// CACHE_TOKEN_TTL. Timeout for each record in cache. Default to 1 h
	TokenTTL time.Duration `mapstructure:"auth_token_ttl" validate:"min=1s,max=1h"`
	// CACHE_CLEANUP_INT. Cache cleanup runs with this time interval. Default to 24 h
	CacheCleanupInterval time.Duration `mapstructure:"auth_cache_cleanup_int" validate:"min=1s,max=24h"`
	// HMAC_SECRET. Secret for token decoding. Required
	HmacSecret string `mapstructure:"auth_hmac_secret" validate:"required,alphanum,min=6,max=32"`
}

type Db struct {
	// DB_DSN. DB dsn. Required
	Dsn string `mapstructure:"db_dsn" validate:"uri"`
	// DB_CONN_MAX. The maximum number of open connections to the database. Default to 5
	ConnMax int `mapstructure:"db_conn_max" validate:"min=1,max=1000"`
	// DB_CONN_MAX_IDLE. The maximum number of connections in the idle connection pool. Default to 5
	ConnMaxIdle int `mapstructure:"db_conn_max_idle" validate:"min=1,max=1000"`
	// DB_CONN_MAX_REUSE. The maximum amount of time a connection may be reused. Default to 1 s
	ConnMaxReuse time.Duration `mapstructure:"db_conn_max_reuse" validate:"min=10ms,max=1h"`
	// DB_DELETE_SESSION_TIMEOUT. The maximum for delete request to run on cache cleaner. Default to 100 ms
	DeleteSessionTimeout time.Duration `mapstructure:"db_delete_session_timeout" validate:"min=10ms,max=1s"`
}

type Config struct {
	Http Http    `mapstructure:",squash"`
	Log  Logging `mapstructure:",squash"`
	Auth Auth    `mapstructure:",squash"`
	Db   Db      `mapstructure:",squash"`
}

func New() (Config, error) {
	c := Config{}

	err := loadEnv(&c)
	if err != nil {
		return Config{}, fmt.Errorf("config unmarshalling error: %w", err)
	}

	err = validator.ValInstance.ValidateStruct(c)
	if err != nil {
		return Config{}, err
	}

	return c, nil
}
