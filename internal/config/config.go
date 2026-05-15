package config

import "time"

type Config struct {
	Server ServerConfig `koanf:"server" json:"server"`
	Log    LogConfig    `koanf:"log" json:"log"`
	Redis  RedisConfig  `koanf:"redis" json:"redis"`
}

type ServerConfig struct {
	Addr            string `koanf:"addr" json:"addr"`
	ShutdownTimeout int    `koanf:"shutdown_timeout" json:"shutdown_timeout"`

	// Resolved at load time — not populated from config directly.
	// Unexported so it isn't accidentally (de)serialised.
	resolvedShutdownTimeout time.Duration
}

type LogConfig struct {
	Level         string              `koanf:"level" json:"level"`
	OpenTelemetry OpenTelemetryConfig `koanf:"opentelemetry" json:"opentelemetry"`
}

type OpenTelemetryConfig struct {
	Enabled bool `koanf:"enabled" json:"enabled"`
}

type RedisConfig struct {
	Addr         string  `koanf:"addr" json:"addr"`
	Username     *string `koanf:"username" json:"username,omitempty"`
	UsernameFile *string `koanf:"username_file" json:"username_file,omitempty"`
	Password     *string `koanf:"password" json:"password,omitempty"`
	PasswordFile *string `koanf:"password_file" json:"password_file,omitempty"`
	DB           int     `koanf:"db" json:"db"`

	// Resolved at load time — not populated from config directly.
	// Unexported so it isn't accidentally (de)serialised.
	resolvedUsername string
	resolvedPassword string
}
