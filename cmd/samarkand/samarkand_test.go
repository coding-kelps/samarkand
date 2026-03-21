package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/coding-kelps/samarkand/internal/config"
)

func successLoader(path string) (config.Config, error) {
	return config.Default(), nil
}

func failingLoader(path string) (config.Config, error) {
	return config.Config{}, errors.New("config file not found")
}

func run(args []string, loader func(string) (config.Config, error)) (string, error) {
	var buf bytes.Buffer
	cmd := newCommand(&buf, loader)
	err := cmd.Run(context.Background(), args)
	return buf.String(), err
}

func TestVersionCommand_PrintsOutput(t *testing.T) {
	out, err := run([]string{"samarkand", "version"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(out) == "" {
		t.Error("expected version output, got empty string")
	}
}

func TestStartCommand_LoadsDefaultConfig(t *testing.T) {
	out, err := run([]string{"samarkand", "start"}, successLoader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Starting with config:") {
		t.Errorf("unexpected output: %q", out)
	}
}

func TestStartCommand_PassesConfigPathToLoader(t *testing.T) {
	wantPath := "/etc/samarkand/config.yaml"
	var gotPath string

	capturingLoader := func(path string) (config.Config, error) {
		gotPath = path
		return config.Default(), nil
	}

	_, err := run([]string{"samarkand", "--config", wantPath, "start"}, capturingLoader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != wantPath {
		t.Errorf("loader received path %q, want %q", gotPath, wantPath)
	}
}

func TestStartCommand_EmptyConfigPath_WhenFlagOmitted(t *testing.T) {
	var gotPath string

	capturingLoader := func(path string) (config.Config, error) {
		gotPath = path
		return config.Default(), nil
	}

	_, err := run([]string{"samarkand", "start"}, capturingLoader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "" {
		t.Errorf("expected empty config path, got %q", gotPath)
	}
}

func TestStartCommand_PropagatesLoaderError(t *testing.T) {
	_, err := run([]string{"samarkand", "start"}, failingLoader)
	if err == nil {
		t.Fatal("expected error from failing loader, got nil")
	}
	if !strings.Contains(err.Error(), "loading config") {
		t.Errorf("error should mention 'loading config', got: %v", err)
	}
}

func TestUnknownSubcommand_ReturnsError(t *testing.T) {
	_, err := run([]string{"samarkand", "notacommand"}, nil)
	if err == nil {
		t.Fatal("expected error for unknown subcommand, got nil")
	}
}
