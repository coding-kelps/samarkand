package config

import (
	"os"
	"strings"
	"time"
)

func (s *ServerConfig) ResolvedShutdownTimeout() *string {
	return nil
}

func (s *ServerConfig) Resolve() error {
	s.resolvedShutdownTimeout = time.Duration(s.ShutdownTimeout) * time.Second

	return nil
}

func (r *RedisConfig) ResolvedUsername() string {
	return r.resolvedUsername
}

func (r *RedisConfig) ResolvedPassword() string {
	return r.resolvedPassword
}

// Resolve validates the username and password configuration and, if
// UsernameFile or PasswordFile is set, reads the file and stores its trimmed
// contents as the effective username or password.
func (r *RedisConfig) Resolve() error {
	if r.UsernameFile != nil {
		data, err := os.ReadFile(*r.UsernameFile)
		if err != nil {
			return err
		}
		s := strings.TrimRight(string(data), "\r\n")
		r.resolvedUsername = s
		return nil
	} else {
		r.resolvedUsername = *r.Username
	}

	if r.PasswordFile != nil {
		data, err := os.ReadFile(*r.PasswordFile)
		if err != nil {
			return err
		}
		s := strings.TrimRight(string(data), "\r\n")
		r.resolvedPassword = s
		return nil
	} else {
		r.resolvedPassword = *r.Password
	}

	return nil
}
