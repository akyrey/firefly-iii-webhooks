package firefly

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifySignature(t *testing.T) {
	tests := []struct {
		name            string
		signatureHeader string
		body            string
		secret          string
		expected        error
	}{
		{
			name:            "empty signature header",
			signatureHeader: "",
			body:            "{\"foo\":\"bar\"}",
			secret:          "abcdef",
			expected:        ErrFireflyInvalidSignature,
		},
		{
			name:            "signature header with invalid format: missing separator",
			signatureHeader: "some-header",
			body:            "{\"foo\":\"bar\"}",
			secret:          "abcdef",
			expected:        ErrFireflyInvalidSignature,
		},
		{
			name:            "signature header with invalid format: missing parts",
			signatureHeader: "some,header",
			body:            "{\"foo\":\"bar\"}",
			secret:          "abcdef",
			expected:        ErrFireflyInvalidSignature,
		},
		{
			name:            "signature header with invalid format: missing time",
			signatureHeader: ",v1=qwerty",
			body:            "{\"foo\":\"bar\"}",
			secret:          "abcdef",
			expected:        ErrFireflyInvalidSignature,
		},
		{
			name:            "signature header with invalid format: missing signature",
			signatureHeader: "t=1610738765,",
			body:            "{\"foo\":\"bar\"}",
			secret:          "abcdef",
			expected:        ErrFireflyInvalidSignature,
		},
		{
			name:            "signature header with invalid format: invalid signature",
			signatureHeader: "t=1610738765,v1=qwerty",
			body:            "{\"foo\":\"bar\"}",
			secret:          "abcdef",
			expected:        ErrFireflyInvalidSignature,
		},
		{
			name:            "signature with empty body",
			signatureHeader: "t=1610738765,v1=de95f8c28fbeab595d5520205a3b7c2a552811573548d4ad6be786c59a69a495",
			body:            "{}",
			secret:          "abcdef",
			expected:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := WebhookMessage{}
			actual := msg.VerifySignature(tt.signatureHeader, tt.body, tt.secret)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
