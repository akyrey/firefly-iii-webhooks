package internal

import (
	"flag"
	"log/slog"
)

// Config holds basic application configuration.
type Config struct {
	Addr              string
	FireflyKey        string
	FireflyConfigFile string
	LogLevel          slog.Level
}

// Parse parses the command line flags and stores the result in the Config struct.
func (c *Config) Parse() {
	flag.StringVar(&c.Addr, "addr", ":4000", "HTTP network address")
	flag.StringVar(&c.FireflyKey, "firefly-key", "", "Firefly III API key to use")
	flag.StringVar(&c.FireflyConfigFile, "firefly-config", "./config.json", "JSON configuration file for Firefly webhooks")
	var logLevel string
	flag.StringVar(&logLevel, "log-level", "debug", "Log message level")
	level, err := parseLogLevel(logLevel)
	if err != nil {
		level = slog.LevelError
	}
	c.LogLevel = level

	flag.Parse()
}

// parseLogLevel converts a string log level to a slog.Level.
func parseLogLevel(s string) (slog.Level, error) {
	var level slog.Level
	err := level.UnmarshalText([]byte(s))
	return level, err
}
