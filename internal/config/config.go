package config

import "time"

type Config struct {
	Server ServerConfig `koanf:"server"`
	Log    LogConfig    `koanf:"log"`
}

type ServerConfig struct {
	GRPCAddr        string        `koanf:"grpc_addr"`
	HTTPAddr        string        `koanf:"http_addr"`
	ShutdownTimeout time.Duration `koanf:"shutdown_timeout"`
}

type LogConfig struct {
	Level string `koanf:"level"`
}

func Default() Config {
	return Config{
		Server: ServerConfig{
			GRPCAddr:        ":50051",
			HTTPAddr:        ":80",
			ShutdownTimeout: 30 * time.Second,
		},
		Log: LogConfig{
			Level: "info",
		},
	}
}
