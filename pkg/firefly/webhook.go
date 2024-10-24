package firefly

// The UUID is unique for each webhook message. You can use it for debug purposes.
// The user ID matches the user who created the webhook, and the trigger + response fields tell you why the webhook
// was fired and what the content of the content field is.
type Webhook struct {
	Content  interface{}     `json:"content"`
	Uuid     string          `json:"uuid"`
	Trigger  WebhookTrigger  `json:"trigger"`
	Response WebhookResponse `json:"response"`
	Url      string          `json:"url"`
	Version  string          `json:"version"`
	UserId   int             `json:"user_id"`
}

type (
	WebhookTrigger  string
	WebhookResponse string
)

const (
	// Triggers
	STORE_TRANSACTION   WebhookTrigger = "STORE_TRANSACTION"
	UPDATE_TRANSACTION  WebhookTrigger = "UPDATE_TRANSACTION"
	DESTROY_TRANSACTION WebhookTrigger = "DESTROY_TRANSACTION"
	// Responses
	RESPONSE_TRANSACTIONS WebhookResponse = "TRANSACTIONS"
	RESPONSE_ACCOUNTS     WebhookResponse = "ACCOUNTS"
	RESPONSE_NONE         WebhookResponse = "NONE"
)

// TODO: add signature verification
