package firefly

import (
	"crypto/hmac"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/akyrey/firefly-iii-webhooks/pkg/firefly/models"
	"golang.org/x/crypto/sha3"
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

// verifySignature will check if the signature is valid for the current message.
// Signature example: t=1610738765,v1=d62463af1dcdcc7b5a2db6cf6b1e01d985c31685ee75d01a4f40754dbb4cf396
func (msg *WebhookMessage) VerifySignature(signatureHeader, body, secret string) error {
	parts := strings.Split(signatureHeader, ",")
	if len(parts) != 2 {
		return ErrFireflyInvalidSignature
	}

	timestampPart := strings.Split(parts[0], "=")
	signaturePart := strings.Split(parts[1], "=")
	if len(timestampPart) != 2 || len(signaturePart) != 2 {
		return ErrFireflyInvalidSignature
	}

	concatenation := fmt.Sprintf("%s.%s", timestampPart[1], body)
	dataHmac := hmac.New(sha3.New256, []byte(secret))
	_, err := dataHmac.Write([]byte(concatenation))
	if err != nil {
		return ErrFireflyInvalidSignature
	}

	if fmt.Sprintf("%x", dataHmac.Sum(nil)) != signaturePart[1] {
		return ErrFireflyInvalidSignature
	}

	return nil
}

type WebhookMessageTransaction struct {
	ID           int                  `json:"id"`
	User         int                  `json:"user"`
	Transactions []models.Transaction `json:"transactions"`
}
