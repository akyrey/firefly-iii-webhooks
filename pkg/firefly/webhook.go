package firefly

import (
	"encoding/json"

	"github.com/akyrey/firefly-iii-webhooks/pkg/firefly/models"
)

// The UUID is unique for each webhook message. You can use it for debug purposes.
// The user ID matches the user who created the webhook, and the trigger + response fields tell you why the webhook
// was fired and what the content of the content field is.
type WebhookMessage struct {
	RawContent json.RawMessage `json:"content"`
	Content    interface{}     `json:"-"`
	Uuid       string          `json:"uuid"`
	Trigger    WebhookTrigger  `json:"trigger"`
	Response   WebhookResponse `json:"response"`
	Url        string          `json:"url"`
	Version    string          `json:"version"`
	UserId     int             `json:"user_id"`
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

func (msg *WebhookMessage) UnmarshalJSON(b []byte) error {
	// INFO: workaround to avoid infinite recursion
	type TmpJson WebhookMessage
	var res TmpJson
	if err := json.Unmarshal(b, &res); err != nil {
		return err
	}

	switch res.Response {
	case RESPONSE_TRANSACTIONS:
		var transaction WebhookMessageTransaction
		if err := json.Unmarshal(res.RawContent, &transaction); err != nil {
			return err
		}
		res.Content = transaction
	case RESPONSE_ACCOUNTS:
	// TODO: implement this type
	// var thing Something2
	// if err := json.Unmarshal(msg.RawContent, &thing); err != nil {
	// 	return err
	// }
	// msg.Content = thing
	case RESPONSE_NONE:
		res.Content = nil
	}

	*msg = WebhookMessage(res)
	return nil
}

type WebhookMessageTransaction struct {
	ID           int                  `json:"id"`
	User         int                  `json:"user"`
	Transactions []models.Transaction `json:"transactions"`
}
