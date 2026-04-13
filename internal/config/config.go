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
	Level string `koanf:"level" json:"level"`
}

type RedisConfig struct {
	Addr         string  `koanf:"addr" json:"addr"`
	Password     *string `koanf:"password" json:"password,omitempty"`
	PasswordFile *string `koanf:"password_file" json:"password_file,omitempty"`
	Db           int     `koanf:"db" json:"db"`

	// Resolved at load time — not populated from config directly.
	// Unexported so it isn't accidentally (de)serialised.
	resolvedPassword *string
}
