package firefly

import "errors"

var (
	ErrFireflyConfigNotFound   = errors.New("configuration not found")
	ErrFireflyEmptyApiKey      = errors.New("api key cannot be empty")
	ErrFireflyInvalidSignature = errors.New("invalid signature")
	ErrFireflyInvalidSecret    = errors.New("invalid signature secret")
)
