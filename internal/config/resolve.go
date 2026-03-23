package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func (r *RedisConfig) ResolvedPassword() *string {
	return r.resolvedPassword
}

// Resolve validates the password configuration and, if PasswordFile is set,
// reads the file and stores its trimmed contents as the effective password.
func (r *RedisConfig) Resolve() error {
	if r.Password != nil && r.PasswordFile != nil {
		return errors.New("redis: password and password_file are mutually exclusive")
	}

	if r.PasswordFile != nil {
		data, err := os.ReadFile(*r.PasswordFile)
		if err != nil {
			return fmt.Errorf("redis: reading password_file: %w", err)
		}
		s := strings.TrimRight(string(data), "\r\n")
		r.resolvedPassword = &s
		return nil
	}

	r.resolvedPassword = r.Password
	return nil
}
