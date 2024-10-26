package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/akyrey/firefly-iii-webhooks/pkg/firefly"
	"github.com/akyrey/firefly-iii-webhooks/pkg/firefly/models"
	"github.com/jinzhu/copier"
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

	// 4. Check content type
	content, ok := webhookMessage.Content.(firefly.WebhookMessageTransaction)
	if !ok {
		app.Logger.Error("Invalid content type")
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	count := len(content.Transactions)
	// 5. Only apply to single transactions and to transactions with foreing amount and currency
	if count == 0 || count > 1 {
		app.Logger.Debug("Found zero or more than one transactions", "count", count)
		app.clientResponse(w, r, http.StatusNoContent)
		return
	}

	t := content.Transactions[0]
	if t.ForeignAmount == nil || t.ForeignCurrencyDecimalPlaces == nil {
		app.Logger.Error("Transactions missing foreign amount info", "transaction", t)
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	// 6. Calculate the foreignAmount to split, using foreignAmount / 8 and foreignAmount % 8
	foreignAmount, err := strconv.ParseFloat(strings.TrimSpace(*t.ForeignAmount), 64)
	if err != nil {
		app.Logger.Error("Invalid foreign amount", "amount", *t.ForeignAmount)
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	zeroWithDelta := math.Pow10(-*t.ForeignCurrencyDecimalPlaces)
	if t.SourceID != config.SourceAccountId || math.Abs(foreignAmount-config.SplitAmount) <= zeroWithDelta {
		app.Logger.Debug("Transaction doesn't meet the requirements", "transaction", t)
		app.clientResponse(w, r, http.StatusNoContent)
		return
	}

	app.Logger.Debug("Transaction meets the requirements", "transaction", t)
	// 8. If the module isn't 0, update this transaction setting the amount to the amount / 8 result
	//    and clone the transaction setting the currency to Satispay and the amount to amount % 8 result
	division := foreignAmount / config.SplitAmount
	updatedAmount := fmt.Sprintf("%.[2]*[1]f", division, t.CurrencyDecimalPlaces)
	updatedForeignAmount := fmt.Sprintf("%.[2]*[1]f", division*config.SplitAmount, t.ForeignCurrencyDecimalPlaces)
	var tToUpdate models.Transaction
	err = copier.Copy(&tToUpdate, &t)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	tToUpdate.Amount = updatedAmount
	tToUpdate.ForeignAmount = &updatedForeignAmount
	tToUpdate.Tags = append(tToUpdate.Tags, "Webhook: split_ticket", fmt.Sprintf("Webhook uuid: %s", webhookMessage.Uuid))
	// TODO: update transaction
	// if err != nil {
	// 	app.serverError(w, r, err)
	// 	return
	// }
	modulo := math.Mod(foreignAmount, config.SplitAmount)
	if modulo <= zeroWithDelta {
		app.clientResponse(w, r, http.StatusNoContent)
		return
	}
	moduloAmount := fmt.Sprintf("%.[2]*[1]f", modulo, config.DestinationCurrencyDecimalPlaces)
	var tToCreate models.Transaction
	err = copier.Copy(&tToCreate, &t)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	tToCreate.Amount = moduloAmount
	tToCreate.CurrencyID = config.DestinationCurrencyId
	tToCreate.ForeignAmount = nil
	tToCreate.ForeignCurrencyID = nil
	tToCreate.Tags = append(tToCreate.Tags, "Webhook: split_ticket", fmt.Sprintf("Webhook uuid: %s", webhookMessage.Uuid))
	// TODO: create transaction
	// if err != nil {
	// 	app.serverError(w, r, err)
	// 	return
	// }
	app.clientResponse(w, r, http.StatusNoContent)
}
