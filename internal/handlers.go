package internal

import (
	"encoding/json"
	"io"
	"math"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/akyrey/firefly-iii-webhooks/pkg/firefly"
)

// splitTicket will split a transaction related to an account into 2 transactions
// each with a different amount and currency as defined in the configuration.
func (app *Application) splitTicket(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// 1. Parse the request body
	var webhookMessage firefly.WebhookMessage
	err = json.Unmarshal(b, &webhookMessage)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	app.Logger.Debug("Received body", "body", webhookMessage)

	// 2. Find the appropriate configuration
	cIdx := slices.IndexFunc(
		app.FireflyConfig.SplitTicket,
		func(c firefly.SplitTicketConfig) bool {
			return c.AppliesTo(webhookMessage)
		},
	)
	if cIdx == -1 {
		app.Logger.Debug("No configuration found")
		app.clientError(w, r, http.StatusNotFound)
		return
	}

	config := app.FireflyConfig.SplitTicket[cIdx]
	app.Logger.Debug("Found configuration", "config", config)
	// TODO:
	// 3. Check if the header contains the signature and verify it

	// 7. Calculate the amount to split, using amount / 8 and amount % 8
	content, ok := webhookMessage.Content.(firefly.WebhookMessageTransaction)
	if !ok {
		app.Logger.Error("Invalid content type")
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	count := len(content.Transactions)
	// Only apply to single transactions
	if count == 0 || count > 1 {
		app.Logger.Debug("Found zero or more than one transactions", "count", count)
		app.clientResponse(w, r, http.StatusNoContent)
		return
	}

	t := content.Transactions[0]
	amount, err := strconv.ParseFloat(strings.TrimSpace(t.Amount), 64)
	if err != nil {
		app.Logger.Error("Invalid amount", "amount", t.Amount)
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	if t.SourceID != config.SourceAccountId ||
		math.Abs(amount-config.SplitAmount) <= math.Pow10(-t.CurrencyDecimalPlaces) {
		app.Logger.Debug("Transaction doesn't meet the requirements", "transaction", t)
		app.clientResponse(w, r, http.StatusNoContent)
		return
	}

	app.Logger.Debug("Transaction meets the requirements", "transaction", t)
	// 8. If the module isn't 0, update this transaction setting the amount to the amount / 8 result
	//    and clone the transaction setting the currency to Satispay and the amount to amount % 8 result
	// updatedAmount := fmt.Sprintf("%.[2]*[1]f", amount/config.SplitAmount, t.CurrencyDecimalPlaces)
	// moduloAmount := fmt.Sprintf("%.[2]*[1]f", math.Mod(amount, config.SplitAmount), config.DestinationCurrencyDecimalPlaces)
}
