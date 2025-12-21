package config

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// Config holds all server configuration
type Config struct {
	// Server settings
	Address     string
	MaxSessions int

	// TUI settings
	Mouse bool
	Paste bool

	// Logging settings
	LogLevel string
}

// Default configuration values
var defaults = Config{
	Address:     ":7755",
	MaxSessions: 100,
	Mouse:       true,
	Paste:       true,
	LogLevel:    "info",
}

// Load loads configuration from file, environment variables, and command-line flags
// Priority: flags > env vars > config file > defaults
func Load() (*Config, error) {
	cfg := defaults

	// 1. Load from config file
	if err := loadFromFile(&cfg, ".yutani.conf"); err != nil {
		// Config file is optional, only log if it exists but has errors
		if !os.IsNotExist(err) {
			slog.Warn("Error loading config file", "error", err)
		}
	}

	// 2. Load from environment variables
	loadFromEnv(&cfg)

	// 3. Load from command-line flags
	loadFromFlags(&cfg)

	return &cfg, nil
}

// loadFromFile reads configuration from a file
func loadFromFile(cfg *Config, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid config line %d: %s", lineNum, line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove YUTANI_ prefix if present
		key = strings.TrimPrefix(key, "YUTANI_")

		if err := setConfigValue(cfg, key, value); err != nil {
			return fmt.Errorf("error on line %d: %w", lineNum, err)
		}
	}

	return scanner.Err()
}

// loadFromEnv reads configuration from environment variables
func loadFromEnv(cfg *Config) {
	if val := os.Getenv("YUTANI_ADDRESS"); val != "" {
		cfg.Address = val
	}
	if val := os.Getenv("YUTANI_MAX_SESSIONS"); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			cfg.MaxSessions = i
		}
	}
	if val := os.Getenv("YUTANI_MOUSE"); val != "" {
		cfg.Mouse = parseBool(val)
	}
	if val := os.Getenv("YUTANI_PASTE"); val != "" {
		cfg.Paste = parseBool(val)
	}
	if val := os.Getenv("YUTANI_LOG_LEVEL"); val != "" {
		cfg.LogLevel = val
	}
}

// loadFromFlags reads configuration from command-line flags
func loadFromFlags(cfg *Config) {
	address := flag.String("address", cfg.Address, "gRPC server listen address")
	maxSessions := flag.Int("max-sessions", cfg.MaxSessions, "Maximum concurrent sessions")
	mouse := flag.Bool("mouse", cfg.Mouse, "Enable mouse support")
	paste := flag.Bool("paste", cfg.Paste, "Enable paste support")
	logLevel := flag.String("log-level", cfg.LogLevel, "Log level (debug, info, warn, error)")

	flag.Parse()

	cfg.Address = *address
	cfg.MaxSessions = *maxSessions
	cfg.Mouse = *mouse
	cfg.Paste = *paste
	cfg.LogLevel = *logLevel
}

// setConfigValue sets a configuration value by key
func setConfigValue(cfg *Config, key, value string) error {
	switch strings.ToUpper(key) {
	case "ADDRESS":
		cfg.Address = value
	case "MAX_SESSIONS":
		i, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid MAX_SESSIONS value: %s", value)
		}
		cfg.MaxSessions = i
	case "MOUSE":
		cfg.Mouse = parseBool(value)
	case "PASTE":
		cfg.Paste = parseBool(value)
	case "LOG_LEVEL":
		cfg.LogLevel = value
	default:
		return fmt.Errorf("unknown config key: %s", key)
	}
	return nil
}

// parseBool parses a boolean value from string
func parseBool(s string) bool {
	s = strings.ToLower(s)
	return s == "true" || s == "1" || s == "yes" || s == "on"
}

