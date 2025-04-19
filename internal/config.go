package internal

import (
	"flag"
	"log/slog"
	"os"
	"strings"
)

// Config holds basic application configuration.
type Config struct {
	Addr              string
	FireflyBaseUrl    string
	FireflyConfigFile string
	FireflyApiKey     string
	LogLevel          slog.Level
}

const (
	// HTTP network address the server listens on.
	ADDRESS = "addr"
	// Log message level.
	LOG_LEVEL = "log-level"
	// Base URL for the Firefly III API.
	BASE_URL = "firefly-base-url"
	// JSON configuration file for Firefly webhooks.
	CONFIG_FILE = "firefly-config"
	// Firefly III API key to use.
	API_KEY = "firefly-api-key"
)

// Parse parses the command line flags and stores the result in the Config struct.
func (c *Config) Parse() {
	parseFlagOrEnv(&c.Addr, ADDRESS, ":4000", "HTTP network address")
	parseFlagOrEnv(&c.FireflyBaseUrl, BASE_URL, "http://firefly_iii_core:8080", "Base URL for the Firefly III API")
	parseFlagOrEnv(&c.FireflyConfigFile, CONFIG_FILE, "./config.json", "JSON configuration file for Firefly webhooks")
	parseFlagOrEnv(&c.FireflyApiKey, API_KEY, "", "Firefly III API key to use")
	var logLevel string
	parseFlagOrEnv(&logLevel, LOG_LEVEL, "debug", "Log message level")
	level, err := parseLogLevel(logLevel)
	if err != nil {
		level = slog.LevelError
	}
	c.LogLevel = level

	flag.Parse()
}

// parseFlagOrEnv parses a flag or an environment variable.
func parseFlagOrEnv(p *string, key, def, description string) {
	flag.StringVar(p, envToFlag(key), getEnvOrDefault(key, def), description)
}

// getEnvOrDefault returns the value of an environment variable or a default value.
func getEnvOrDefault(key, def string) string {
	env := os.Getenv(flagToEnv(key))
	if env == "" {
		return def
	}

	return env
}

// parseLogLevel converts a string log level to a slog.Level.
func parseLogLevel(s string) (slog.Level, error) {
	var level slog.Level
	err := level.UnmarshalText([]byte(s))
	return level, err
}

// envToFlag converts THIS_FORMAT to this-format.
func envToFlag(e string) string {
	return strings.ReplaceAll(strings.ToLower(e), "_", "-")
}

// flagToEnv converts this-format to THIS_FORMAT.
func flagToEnv(f string) string {
	return strings.ReplaceAll(strings.ToUpper(f), "-", "_")
}
