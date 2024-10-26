package main

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/akyrey/firefly-iii-webhooks/internal"
	"github.com/akyrey/firefly-iii-webhooks/pkg/assert"
	"github.com/akyrey/firefly-iii-webhooks/pkg/firefly"
	"github.com/akyrey/firefly-iii-webhooks/pkg/prettylog"
)

func main() {
	config := internal.Config{}
	config.Parse()

	logger := slog.New(prettylog.NewHandler(&slog.HandlerOptions{
		AddSource: true,
		Level:     config.LogLevel,
	}))

	app := &internal.Application{
		Config: config,
		FireflyClient: firefly.NewFirefly(
			config.FireflyBaseUrl,
			firefly.WithApiKey(config.FireflyKey),
		),
		FireflyConfig: firefly.ReadConfig(config.FireflyConfigFile),
		Logger:        logger,
	}

	srv := &http.Server{
		Addr:    config.Addr,
		Handler: app.Routes(config),
		// Server-wide settings which act on the underlying connection.
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	logger.Info("starting server", "addr", srv.Addr)

	err := srv.ListenAndServe()
	assert.NoError(err, "server failed to start", "error", err)
}
