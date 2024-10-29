package firefly

import "errors"

var (
	ErrFireflyEmptyApiKey      = errors.New("api key cannot be empty")
	ErrFireflyInvalidSignature = errors.New("invalid signature")
	ErrFireflyInvalidSecret    = errors.New("invalid signature secret")
)
