package config

import "time"

type Config struct {
	Server ServerConfig `koanf:"server"`
	Log    LogConfig    `koanf:"log"`
	Redis RedisConfig `koanf:"redis"`
}

type ServerConfig struct {
	Addr            string        `koanf:"addr"`
	ShutdownTimeout time.Duration `koanf:"shutdown_timeout"`
}

type LogConfig struct {
	Level string `koanf:"level"`
}

type RedisConfig struct {
	Addr string `koanf:"addr"`
	Password *string `koanf:"password"`
	PasswordFile *string `koanf:"password_file"`
	Db string `koanf:"db"`

	// Resolved at load time — not populated from config directly.
	// Unexported so it isn't accidentally (de)serialised.
	resolvedPassword *string
}

func Default() Config {
	return Config{
		Server: ServerConfig{
			Addr:            ":50051",
			ShutdownTimeout: 30 * time.Second,
		},
		Log: LogConfig{
			Level: "info",
		},
		Redis: RedisConfig{
			Addr: "localhost:6739",
			Password: nil,
			PasswordFile: nil,
			Db: "0",
		},
	}
}
