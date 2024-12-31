package config

import (
	"github.com/spf13/viper"
)

func setNetworkEnv() {
	viper.SetDefault("net_host", "0.0.0.0")
	_ = viper.BindEnv("net_host")

	viper.SetDefault("net_port", "8080")
	_ = viper.BindEnv("net_port")

	viper.SetDefault("net_read_timeout", "500ms")
	_ = viper.BindEnv("net_read_timeout")

	viper.SetDefault("net_write_timeout", "500ms")
	_ = viper.BindEnv("net_write_timeout")

	viper.SetDefault("net_idle_timeout", "1s")
	_ = viper.BindEnv("net_idle_timeout")
}

func setLoggingEnv() {
	viper.SetDefault("log_format", "json")
	_ = viper.BindEnv("log_format")

	viper.SetDefault("log_level", "info")
	_ = viper.BindEnv("log_level")
}

func setAuthEnv() {
	viper.SetDefault("auth_token_ttl", "1h")
	_ = viper.BindEnv("auth_token_ttl")

	viper.SetDefault("auth_cache_cleanup_int", "24h")
	_ = viper.BindEnv("auth_cache_cleanup_int")

	_ = viper.BindEnv("auth_hmac_secret")
}

func setDbEnv() {
	_ = viper.BindEnv("db_dsn")

	viper.SetDefault("db_conn_max", "5")
	_ = viper.BindEnv("db_conn_max")

	viper.SetDefault("db_conn_max_idle", "5")
	_ = viper.BindEnv("db_conn_max_idle")

	viper.SetDefault("db_conn_max_reuse", "1s")
	_ = viper.BindEnv("db_conn_max_reuse")

	viper.SetDefault("db_delete_session_timeout", "100ms")
	_ = viper.BindEnv("db_delete_session_timeout")
}

func loadEnv(c *Config) error {
	setNetworkEnv()
	setLoggingEnv()
	setAuthEnv()
	setDbEnv()

	viper.AutomaticEnv()
	return viper.Unmarshal(c)
}
