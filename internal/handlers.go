package internal

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/akyrey/firefly-iii-webhooks/pkg/firefly"
)

// splitTicket will split a transaction related to an account into 2 transactions
// each with a different amount and currency as defined in the configuration.
func (app *Application) splitTicket(w http.ResponseWriter, r *http.Request) {
	body, webhookMessage, err := app.parseRequestMessage(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	configValue, err := app.FireflyConfig.FindConfig(firefly.SplitTicket, webhookMessage)
	if err != nil {
		app.Logger.Debug("No configuration found", "error", err)
		app.clientError(w, r, http.StatusNotFound)
		return
	}
	config, ok := configValue.(firefly.SplitTicketConfig)
	if !ok {
		app.Logger.Error("Invalid configuration type", "config", configValue)
		app.clientError(w, r, http.StatusInternalServerError)
		return
	}
	app.Logger.Debug("Found configuration", "config", config)

	app.Logger.Debug("Verifying signature", "signature", r.Header.Get("Signature"))
	err = webhookMessage.VerifySignature(r.Header.Get("Signature"), string(body), config.Secret)
	if err != nil {
		app.Logger.Error("Failed validating signature", "header", r.Header.Get("Signature"), "error", err)
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	content, ok := webhookMessage.Content.(firefly.WebhookMessageTransaction)
	if !ok {
		app.Logger.Error("Invalid content type", "content", webhookMessage.Content)
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	count := len(content.Transactions)
	// Only apply to single transactions and to transactions with foreing amount and currency
	if count == 0 || count > 1 {
		app.Logger.Debug("Found zero or more than one transactions", "count", count)
		app.clientResponse(w, r, http.StatusNoContent)
		return
	}

	t := content.Transactions[0]
	if t.SourceID != config.SourceAccountId {
		app.Logger.Debug("Transaction source id different from configured one", "transaction", t)
		app.clientResponse(w, r, http.StatusNoContent)
		return
	}
	if t.ForeignAmount == nil || t.ForeignCurrencyDecimalPlaces == nil {
		app.Logger.Error("Transactions missing foreign amount info", "transaction", t)
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	foreignAmount, err := strconv.ParseFloat(strings.TrimSpace(*t.ForeignAmount), 64)
	if err != nil {
		app.Logger.Error("Invalid foreign amount", "amount", *t.ForeignAmount)
		app.clientError(w, r, http.StatusBadRequest)
		return
	}

	app.Logger.Debug("Transaction meets the requirements", "transaction", t)
	zeroWithDelta := math.Pow10(-*t.ForeignCurrencyDecimalPlaces)
	division := math.Floor(foreignAmount / config.SplitAmount)
	if division <= zeroWithDelta {
		app.Logger.Debug("No need to update the transaction: division lesser than zero", "division", division)
		app.clientResponse(w, r, http.StatusNoContent)
		return
	}
	// Update this transaction setting the amount to the amount / config.SplitAmount result
	err = app.updateSplitTransaction(&t, content.ID, webhookMessage.Uuid, division, config.SplitAmount)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	modulo := math.Mod(foreignAmount, config.SplitAmount)
	if modulo <= zeroWithDelta {
		app.Logger.Debug("No need to create new transaction: remainder lesser than zero", "modulo", modulo)
		app.clientResponse(w, r, http.StatusNoContent)
		return
	}
	// If the module isn't 0, create a new transaction with the module amount
	err = app.createSplitTransaction(
		&t,
		webhookMessage.Uuid,
		modulo,
		config.DestinationCurrencyDecimalPlaces,
		config.DestinationAccountId,
		config.DestinationCurrencyId,
	)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	app.Logger.Debug("Webhook completed successfully")
	app.clientResponse(w, r, http.StatusNoContent)
}
