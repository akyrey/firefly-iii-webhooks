package internal

import (
	"flag"
	"log/slog"
)

type Config struct {
	Addr       string
	FireflyKey string
	LogLevel   slog.Level
}

func (c *Config) Parse() {
	flag.StringVar(&c.Addr, "addr", ":4000", "HTTP network address")
	flag.StringVar(&c.FireflyKey, "firefly-key", "", "Firefly III API key to use")
	var logLevel string
	flag.StringVar(&logLevel, "log-level", "debug", "Log message level")
	level, err := parseLogLevel(logLevel)
	if err != nil {
		level = slog.LevelError
	}
	c.LogLevel = level

	flag.Parse()
}

func parseLogLevel(s string) (slog.Level, error) {
	var level slog.Level
	err := level.UnmarshalText([]byte(s))
	return level, err
}
