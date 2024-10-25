package firefly

import (
	"encoding/json"
	"os"

	"github.com/akyrey/firefly-iii-webhooks/pkg/assert"
)

// Config holds configuration regarding Firefly webhooks.
type Config struct {
	SplitTicket []SplitTicketConfig `json:"split_ticket,omitempty"`
}

// SplitTicketConfig holds configuration for splitting a transaction.
type SplitTicketConfig struct {
	Trigger                          WebhookTrigger  `json:"trigger"`
	Response                         WebhookResponse `json:"response"`
	Secret                           string          `json:"secret"`
	Type                             TransactionType `json:"type"`
	SourceAccountId                  int             `json:"source_account_id"`
	DestinationAccountId             int             `json:"destination_account_id"`
	DestinationCurrencyId            int             `json:"destination_currency_id"`
	DestinationCurrencyDecimalPlaces int             `json:"destination_currency_decimal_places"`
	SplitAmount                      float64         `json:"split_amount"`
}

// AppliesTo checks if the configuration applies to the given message.
func (c SplitTicketConfig) AppliesTo(msg WebhookMessage) bool {
	return c.Trigger == msg.Trigger &&
		c.Response == msg.Response &&
		c.Type == WITHDRAWAL &&
		c.SourceAccountId != c.DestinationAccountId
}

// ReadConfig reads the configuration from a JSON file.
func ReadConfig(file string) Config {
	configFile, err := os.Open(file)
	defer configFile.Close()
	assert.NoError(err, "Firefly configuration file should always be provided")

	var config Config
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	assert.NoError(err, "Unable to parse Firefly configuration file")

	return config
}

type TransactionType string

const (
	WITHDRAWAL TransactionType = "withdrawal"
	DEPOSIT    TransactionType = "deposit"
	TRANSFER   TransactionType = "transfer"
)
