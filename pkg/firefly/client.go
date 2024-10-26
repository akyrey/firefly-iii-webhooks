package firefly

import (
	"net/http"
	"strings"
	"time"

	"github.com/akyrey/firefly-iii-webhooks/pkg/assert"
)

// Firefly client used to interact with the Firefly III API.
type Firefly struct {
	httpClient *http.Client
	baseUrl    string
	// Optional configuration options
	fireflyOpts
}

// Create a new Firefly with the given configuration.
func NewFirefly(baseUrl string, opts ...FireflyOption) *Firefly {
	var options fireflyOpts
	for _, opt := range opts {
		err := opt(&options)
		assert.NoError(err, "Error applying Firefly option")
	}

	if options.timeout == 0 {
		options.timeout = defaultTimeout
	}

	return &Firefly{
		baseUrl: baseUrl,
		httpClient: &http.Client{
			Timeout: options.timeout,
		},
		fireflyOpts: options,
	}
}

const defaultTimeout = 10 * time.Second

type fireflyOpts struct {
	apiKey  *string
	timeout time.Duration
}

// FireflyOption is a function that updates the fireflyOpts struct.
type FireflyOption func(*fireflyOpts) error

// WithApiKey is a configuration function that updates the api key used for each request.
func WithApiKey(apiKey string) FireflyOption {
	return func(c *fireflyOpts) error {
		trim := strings.TrimSpace(apiKey)
		if trim == "" {
			return ErrFireflyEmptyApiKey
		}
		c.apiKey = &trim
		return nil
	}
}

// WithTimeout is a configuration function that updates the client timeout.
func WithTimeout(timeout time.Duration) FireflyOption {
	return func(c *fireflyOpts) error {
		c.timeout = timeout
		return nil
	}
}
