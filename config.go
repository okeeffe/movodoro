package main

import (
	"os"
	"path/filepath"
)

// Config holds configuration for the application
type Config struct {
	LogsDir     string
	CurrentPath string
	MovosDir    string
	MaxDailyRPE int
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}

	// Check for MOVODORO_MOVOS_DIR environment variable
	movosDir := os.Getenv("MOVODORO_MOVOS_DIR")
	if movosDir == "" {
		// Fall back to ~/.movodoro/movos
		movosDir = filepath.Join(home, ".movodoro", "movos")
	}

	return &Config{
		LogsDir:     filepath.Join(home, ".movodoro", "logs"),
		CurrentPath: filepath.Join(home, ".movodoro", "current"),
		MovosDir:    movosDir,
		MaxDailyRPE: 30,
	}
}

// TestConfig returns a configuration for testing
func TestConfig(testDir string) *Config {
	return &Config{
		LogsDir:     filepath.Join(testDir, "logs"),
		CurrentPath: filepath.Join(testDir, "current"),
		MovosDir:    filepath.Join(testDir, "test-movos"),
		MaxDailyRPE: 30,
	}
}
