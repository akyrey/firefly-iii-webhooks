package firefly

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func splitTicketConfigJSON(amounts string) []byte {
	return []byte(fmt.Sprintf(`{"split_ticket":[{
		"trigger":"STORE_TRANSACTION","response":"TRANSACTIONS","secret":"s","type":"withdrawal",
		"source_account_id":"1","destination_account_id":"4","destination_currency_id":"1",
		"destination_currency_decimal_places":2,"link_type_id":"1",%s}]}`, amounts))
}

func TestConfigUnmarshalsSplitTicketAmounts(t *testing.T) {
	// Arrange
	raw := splitTicketConfigJSON(`"split_amounts":[8,10]`)

	// Act
	var config Config
	err := json.Unmarshal(raw, &config)

	// Assert
	require.NoError(t, err)
	require.Len(t, config[SplitTicket], 1)
	splitTicket, ok := config[SplitTicket][0].(SplitTicketConfig)
	require.True(t, ok)
	assert.Equal(t, []float64{8, 10}, splitTicket.SplitAmounts)
}

func TestConfigRejectsInvalidSplitTicketAmounts(t *testing.T) {
	tests := []struct {
		name    string
		amounts string
	}{
		{name: "legacy split_amount key", amounts: `"split_amount":8`},
		{name: "empty list", amounts: `"split_amounts":[]`},
		{name: "zero amount", amounts: `"split_amounts":[8,0]`},
		{name: "negative amount", amounts: `"split_amounts":[-8]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			var config Config
			err := json.Unmarshal(splitTicketConfigJSON(tt.amounts), &config)

			// Assert
			assert.ErrorIs(t, err, ErrInvalidSplitAmounts)
		})
	}
}
