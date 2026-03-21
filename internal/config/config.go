package config

import "time"

type Config struct {
	Server ServerConfig `koanf:"server"`
	Log    LogConfig    `koanf:"log"`
}

type ServerConfig struct {
	Addr            string        `koanf:"addr"`
	ShutdownTimeout time.Duration `koanf:"shutdown_timeout"`
}

type LogConfig struct {
	Level string `koanf:"level"`
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
	}
}
