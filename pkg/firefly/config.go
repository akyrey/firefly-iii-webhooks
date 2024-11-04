package firefly

import (
	"encoding/json"
	"os"
	"slices"

	"github.com/akyrey/firefly-iii-webhooks/pkg/assert"
)

// ConfigType is an enum listing all possible configuration types.
type ConfigType string

const (
	SplitTicket ConfigType = "split_ticket"
)

// Config holds configuration regarding Firefly webhooks.
type Config map[ConfigType][]ConfigValue

type ConfigValue interface {
	// AppliesTo checks if the configuration applies to the given message.
	AppliesTo(msg WebhookMessage) bool
}

// UnmarshalJSON unmarshals the JSON configuration file into the Config struct.
func (c *Config) UnmarshalJSON(b []byte) error {
	if *c == nil {
		*c = make(map[ConfigType][]ConfigValue)
	}
	var config map[ConfigType][]json.RawMessage
	if err := json.Unmarshal(b, &config); err != nil {
		return err
	}
	for t, list := range config {
		switch t {
		case SplitTicket:
			var splitTicketList []ConfigValue
			for _, raw := range list {
				var splitTicket SplitTicketConfig
				if err := json.Unmarshal(raw, &splitTicket); err != nil {
					return err
				}
				splitTicketList = append(splitTicketList, splitTicket)
			}
			(*c)[t] = splitTicketList
		}
	}
	return nil
}

// FindConfig finds the configuration that applies to the given message.
func (c *Config) FindConfig(t ConfigType, msg WebhookMessage) (ConfigValue, error) {
	list, ok := (*c)[t]
	if !ok {
		return nil, ErrFireflyConfigNotFound
	}
	cIdx := slices.IndexFunc(
		list,
		func(c ConfigValue) bool {
			return c.AppliesTo(msg)
		},
	)
	if cIdx == -1 {
		return nil, ErrFireflyConfigNotFound
	}

	return list[cIdx], nil
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
func ReadConfig(file string) *Config {
	configFile, err := os.Open(file)
	assert.NoError(err, "Firefly configuration file should always be provided")
	defer configFile.Close()

	var config Config
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	assert.NoError(err, "Unable to parse Firefly configuration file")

	return &config
}

type TransactionType string

const (
	WITHDRAWAL TransactionType = "withdrawal"
	DEPOSIT    TransactionType = "deposit"
	TRANSFER   TransactionType = "transfer"
)
