package internal

import (
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
	if config.SplitAmount == 0 {
		app.Logger.Debug("Invalid split amount", "amount", config.SplitAmount)
		app.clientError(w, r, http.StatusBadRequest)
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
	if count != 1 {
		app.Logger.Debug("Found zero or more than one transactions", "count", count)
		app.clientResponse(w, r, http.StatusNoContent)
		return
	}

	t := content.Transactions[0]
	if t.SourceID != config.SourceAccountId {
		app.Logger.Debug("Transaction source id different from configured one", "transaction", t, "config", config)
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
	updated, err := app.updateSplitTransaction(&t, content.ID, division, config.SplitAmount)
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
	created, err := app.createSplitTransaction(
		&t,
		modulo,
		config.DestinationCurrencyDecimalPlaces,
		config.DestinationAccountId,
		config.DestinationCurrencyId,
	)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	err = app.FireflyClient.LinkTransactions(config.LinkTypeId, updated.Data.ID, created.Data.ID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	app.Logger.Debug("Webhook completed successfully")
	app.clientResponse(w, r, http.StatusNoContent)
}

// cashback will create a new deposit transaction with a static amount
// each with a different amount and currency as defined in the configuration.
func (app *Application) cashback(w http.ResponseWriter, r *http.Request) {
	body, webhookMessage, err := app.parseRequestMessage(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	configValue, err := app.FireflyConfig.FindConfig(firefly.Cashback, webhookMessage)
	if err != nil {
		app.Logger.Debug("No configuration found", "error", err)
		app.clientError(w, r, http.StatusNotFound)
		return
	}
	config, ok := configValue.(firefly.CashbackConfig)
	if !ok {
		app.Logger.Error("Invalid configuration type", "config", configValue)
		app.clientError(w, r, http.StatusInternalServerError)
		return
	}

	if config.Amount <= 0 {
		app.Logger.Debug("Invalid configured amount", "amount", config.Amount)
		app.clientError(w, r, http.StatusBadRequest)
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

	var transactionIDToLink *string
	for _, t := range content.Transactions {
		if t.SourceID != config.SourceAccountId {
			app.Logger.Debug("Transactions source id different from configured one", "transaction", t, "config", config)
			app.clientResponse(w, r, http.StatusNoContent)
			return
		}
		if !slices.Contains(t.Tags, config.SourceMustHaveTag) {
			continue
		}
		created, err := app.createCashbackTransaction(
			&t,
			config,
		)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		transactionIDToLink = &created.Data.ID
	}

	if transactionIDToLink != nil {
		err = app.FireflyClient.LinkTransactions(config.LinkTypeId, strconv.Itoa(content.ID), *transactionIDToLink)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
	}

	app.Logger.Debug("Webhook completed successfully")
	app.clientResponse(w, r, http.StatusNoContent)
}

// transfer will create a new transfer transaction from a source account to a destination account with an amount
// defined by the transaction triggering the webhook.
func (app *Application) transfer(w http.ResponseWriter, r *http.Request) {
	body, webhookMessage, err := app.parseRequestMessage(r)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	configValue, err := app.FireflyConfig.FindConfig(firefly.Transfer, webhookMessage)
	if err != nil {
		app.Logger.Debug("No configuration found", "error", err)
		app.clientError(w, r, http.StatusNotFound)
		return
	}
	config, ok := configValue.(firefly.TransferConfig)
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

	var transactionIDToLink *string
	for _, t := range content.Transactions {
		sourceID := t.SourceID
		// If it's a deposit, the source id is the transaction destination id
		if config.Type == firefly.DEPOSIT {
			sourceID = t.DestinationID
		}
		if sourceID != config.SourceAccountId {
			app.Logger.Debug("Transactions source id different from configured one", "transaction", t, "config", config)
			app.clientResponse(w, r, http.StatusNoContent)
			return
		}
		if !slices.Contains(t.Tags, config.SourceMustHaveTag) {
			continue
		}
		amount := 0.0
		zeroWithDelta := math.Pow10(-t.CurrencyDecimalPlaces)
		if config.FixedAmount != nil {
			amount = *config.FixedAmount
		} else if config.ModuloAmount != nil {
			transactionAmount, err := strconv.ParseFloat(strings.TrimSpace(t.Amount), 64)
			if err != nil {
				app.Logger.Error("Invalid transaction amount", "amount", t.Amount)
				app.clientError(w, r, http.StatusBadRequest)
				return
			}
			amount = *config.ModuloAmount - math.Mod(transactionAmount, *config.ModuloAmount)
		}
		if amount <= zeroWithDelta {
			app.Logger.Debug("No need to create new transaction: remainder lesser than zero", "modulo", amount)
			continue
		}
		created, err := app.createTransferTransaction(
			&t,
			amount,
			config,
		)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
		transactionIDToLink = &created.Data.ID
	}

	if transactionIDToLink != nil {
		err = app.FireflyClient.LinkTransactions(config.LinkTypeId, strconv.Itoa(content.ID), *transactionIDToLink)
		if err != nil {
			app.serverError(w, r, err)
			return
		}
	}

	app.Logger.Debug("Webhook completed successfully")
	app.clientResponse(w, r, http.StatusNoContent)
}
