package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/coding-kelps/samarkand/internal/config"
)

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

// TestStartCommand_LoadsConfig verifies that the start command invokes the
// config loader. A loader that immediately returns an error is used so that
// execution stops after the loader call, before any OTel / server setup that
// would fail in a unit-test environment.
func TestStartCommand_LoadsConfig(t *testing.T) {
	called := false
	sentinelErr := errors.New("stop after load")

	loader := func(_ string) (config.Config, error) {
		called = true
		return config.Config{}, sentinelErr
	}

	_, err := run([]string{"samarkand", "start"}, loader)
	if !called {
		t.Fatal("expected config loader to be called, but it was not")
	}
	if !errors.Is(err, sentinelErr) {
		t.Errorf("expected sentinel error, got: %v", err)
	}
}

func TestStartCommand_PassesConfigPathToLoader(t *testing.T) {
	wantPath := "/etc/samarkand/config.yaml"
	var gotPath string

	// Return an error so the command halts after the loader call, before
	// any OTel / server code that would fail in a test environment.
	capturingLoader := func(path string) (config.Config, error) {
		gotPath = path
		return config.Config{}, errors.New("stop after load")
	}

	_, err := run([]string{"samarkand", "--config", wantPath, "start"}, capturingLoader)
	if err == nil {
		t.Fatal("expected loader error to propagate, got nil")
	}
	if gotPath != wantPath {
		t.Errorf("loader received path %q, want %q", gotPath, wantPath)
	}
}

func TestStartCommand_EmptyConfigPath_WhenFlagOmitted(t *testing.T) {
	var gotPath string

	// Same strategy: bail out early via an error to avoid OTel / server setup.
	capturingLoader := func(path string) (config.Config, error) {
		gotPath = path
		return config.Config{}, errors.New("stop after load")
	}

	_, err := run([]string{"samarkand", "start"}, capturingLoader)
	if err == nil {
		t.Fatal("expected loader error to propagate, got nil")
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
	// The start action returns the raw loader error, so check for the
	// message from failingLoader rather than any wrapper text.
	if !strings.Contains(err.Error(), "config file not found") {
		t.Errorf("expected error to contain 'config file not found', got: %v", err)
	}
}
