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
		{
			name:            "signature real example",
			signatureHeader: "t=1730392952,v1=cfec5771187aa197f412796ba2284897c2514326227845ddbac0253ebf746e25",
			body:            "{\"uuid\":\"23a7fd16-3a55-4cef-85ae-059c666520b7\",\"user_id\":1,\"trigger\":\"STORE_TRANSACTION\",\"response\":\"TRANSACTIONS\",\"url\":\"http:\\/\\/firefly_webhooks:4000\\/api\\/v1\\/webhook\\/split-ticket\",\"version\":\"v0\",\"content\":{\"id\":27,\"created_at\":\"2024-10-31T17:37:37+01:00\",\"updated_at\":\"2024-10-31T17:38:43+01:00\",\"user\":1,\"group_title\":\"\",\"transactions\":[{\"user\":1,\"transaction_journal_id\":27,\"type\":\"withdrawal\",\"date\":\"2024-10-31T17:37:00+01:00\",\"order\":0,\"currency_id\":26,\"currency_code\":\"TKT\",\"currency_symbol\":\"@\",\"currency_decimal_places\":0,\"foreign_currency_id\":1,\"foreign_currency_code\":\"EUR\",\"foreign_currency_symbol\":\"\\u20ac\",\"foreign_currency_decimal_places\":2,\"amount\":\"3\",\"foreign_amount\":\"24.00\",\"description\":\"Testing signature\",\"source_id\":1,\"source_name\":\"Ticket Restaurant\",\"source_iban\":\"\",\"source_type\":\"Asset account\",\"destination_id\":6,\"destination_name\":\"Test\",\"destination_iban\":null,\"destination_type\":\"Expense account\",\"budget_id\":null,\"budget_name\":null,\"category_id\":null,\"category_name\":null,\"bill_id\":null,\"bill_name\":null,\"reconciled\":false,\"notes\":null,\"tags\":[\"Webhook uuid: 2141b03c-c764-42eb-8c69-3f4d65b5a40d\",\"Webhook uuid: 955d0bcc-e641-4f12-8b05-aedf2c9fcfb9\",\"Webhook: split_ticket\"],\"internal_reference\":null,\"external_id\":null,\"original_source\":\"ff3-v6.1.21|api-v2.1.0\",\"recurrence_id\":null,\"bunq_payment_id\":null,\"import_hash_v2\":\"c260533de04f0cd4191c829232c7c6ddece03d32df1906bfad7830be7f8b9728\",\"sepa_cc\":null,\"sepa_ct_op\":null,\"sepa_ct_id\":null,\"sepa_db\":null,\"sepa_country\":null,\"sepa_ep\":null,\"sepa_ci\":null,\"sepa_batch_id\":null,\"interest_date\":null,\"book_date\":null,\"process_date\":null,\"due_date\":null,\"payment_date\":null,\"invoice_date\":null,\"longitude\":null,\"latitude\":null,\"zoom_level\":null}],\"links\":[{\"rel\":\"self\",\"uri\":\"\\/transactions\\/27\"}]}}",
			secret:          "M6FhOxcsfIf5AWo63duzNpCX",
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
