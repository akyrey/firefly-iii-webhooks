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
func (a *Application) splitTicket(w http.ResponseWriter, r *http.Request) {
	body, webhookMessage, err := a.parseRequestMessage(r)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	configValue, err := a.FireflyConfig.FindConfig(firefly.SplitTicket, webhookMessage)
	if err != nil {
		a.Logger.Debug("No configuration found", "error", err)
		a.clientError(w, r, http.StatusNotFound)
		return
	}
	config, ok := configValue.(firefly.SplitTicketConfig)
	if !ok {
		a.Logger.Error("Invalid configuration type", "config", configValue)
		a.clientError(w, r, http.StatusInternalServerError)
		return
	}
	if config.SplitAmount == 0 {
		a.Logger.Debug("Invalid split amount", "amount", config.SplitAmount)
		a.clientError(w, r, http.StatusBadRequest)
		return
	}
	a.Logger.Debug("Found configuration", "config", config)

	a.Logger.Debug("Verifying signature", "signature", r.Header.Get("Signature"))
	err = webhookMessage.VerifySignature(r.Header.Get("Signature"), string(body), config.Secret)
	if err != nil {
		a.Logger.Error("Failed validating signature", "header", r.Header.Get("Signature"), "error", err)
		a.clientError(w, r, http.StatusBadRequest)
		return
	}

	content, ok := webhookMessage.Content.(firefly.WebhookMessageTransaction)
	if !ok {
		a.Logger.Error("Invalid content type", "content", webhookMessage.Content)
		a.clientError(w, r, http.StatusBadRequest)
		return
	}

	count := len(content.Transactions)
	// Only apply to single transactions and to transactions with foreing amount and currency
	if count != 1 {
		a.Logger.Debug("Found zero or more than one transactions", "count", count)
		a.clientResponse(w, r, http.StatusNoContent)
		return
	}

	t := content.Transactions[0]
	if t.SourceID != config.SourceAccountId {
		a.Logger.Debug("Transaction source id different from configured one", "transaction", t, "config", config)
		a.clientResponse(w, r, http.StatusNoContent)
		return
	}
	if t.ForeignAmount == nil || t.ForeignCurrencyDecimalPlaces == nil {
		a.Logger.Error("Transactions missing foreign amount info", "transaction", t)
		a.clientError(w, r, http.StatusBadRequest)
		return
	}

	foreignAmount, err := strconv.ParseFloat(strings.TrimSpace(*t.ForeignAmount), 64)
	if err != nil {
		a.Logger.Error("Invalid foreign amount", "amount", *t.ForeignAmount)
		a.clientError(w, r, http.StatusBadRequest)
		return
	}

	a.Logger.Debug("Transaction meets the requirements", "transaction", t)
	zeroWithDelta := math.Pow10(-*t.ForeignCurrencyDecimalPlaces)
	division := math.Floor(foreignAmount / config.SplitAmount)
	if division <= zeroWithDelta {
		a.Logger.Debug("No need to update the transaction: division lesser than zero", "division", division)
		a.clientResponse(w, r, http.StatusNoContent)
		return
	}
	// Update this transaction setting the amount to the amount / config.SplitAmount result
	updated, err := a.updateSplitTransaction(&t, content.ID, division, config.SplitAmount)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	modulo := math.Mod(foreignAmount, config.SplitAmount)
	if modulo <= zeroWithDelta {
		a.Logger.Debug("No need to create new transaction: remainder lesser than zero", "modulo", modulo)
		a.clientResponse(w, r, http.StatusNoContent)
		return
	}
	// If the module isn't 0, create a new transaction with the module amount
	created, err := a.createSplitTransaction(
		&t,
		modulo,
		config.DestinationCurrencyDecimalPlaces,
		config.DestinationAccountId,
		config.DestinationCurrencyId,
	)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	a.Logger.Debug("Linking transactions", "initial id", updated.Data.ID, "created id", created.Data.ID, "link type", config.LinkTypeId)
	err = a.FireflyClient.LinkTransactions(config.LinkTypeId, updated.Data.ID, created.Data.ID)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	a.Logger.Debug("Webhook completed successfully")
	a.clientResponse(w, r, http.StatusNoContent)
}

// cashback will create a new deposit transaction with a static amount
// each with a different amount and currency as defined in the configuration.
func (a *Application) cashback(w http.ResponseWriter, r *http.Request) {
	body, webhookMessage, err := a.parseRequestMessage(r)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	configValue, err := a.FireflyConfig.FindConfig(firefly.Cashback, webhookMessage)
	if err != nil {
		a.Logger.Debug("No configuration found", "error", err)
		a.clientError(w, r, http.StatusNotFound)
		return
	}
	config, ok := configValue.(firefly.CashbackConfig)
	if !ok {
		a.Logger.Error("Invalid configuration type", "config", configValue)
		a.clientError(w, r, http.StatusInternalServerError)
		return
	}

	if config.Amount <= 0 {
		a.Logger.Debug("Invalid configured amount", "amount", config.Amount)
		a.clientError(w, r, http.StatusBadRequest)
		return
	}
	a.Logger.Debug("Found configuration", "config", config)

	a.Logger.Debug("Verifying signature", "signature", r.Header.Get("Signature"))
	err = webhookMessage.VerifySignature(r.Header.Get("Signature"), string(body), config.Secret)
	if err != nil {
		a.Logger.Error("Failed validating signature", "header", r.Header.Get("Signature"), "error", err)
		a.clientError(w, r, http.StatusBadRequest)
		return
	}

	content, ok := webhookMessage.Content.(firefly.WebhookMessageTransaction)
	if !ok {
		a.Logger.Error("Invalid content type", "content", webhookMessage.Content)
		a.clientError(w, r, http.StatusBadRequest)
		return
	}

	var transactionIDToLink *string
	for _, t := range content.Transactions {
		if t.SourceID != config.SourceAccountId {
			a.Logger.Debug("Transactions source id different from configured one", "transaction", t, "config", config)
			a.clientResponse(w, r, http.StatusNoContent)
			return
		}
		if !slices.Contains(t.Tags, config.SourceMustHaveTag) {
			continue
		}
		created, err2 := a.createCashbackTransaction(
			&t,
			config,
		)
		if err2 != nil {
			a.serverError(w, r, err2)
			return
		}
		transactionIDToLink = &created.Data.ID
	}

	if transactionIDToLink != nil {
		err = a.FireflyClient.LinkTransactions(config.LinkTypeId, strconv.Itoa(content.ID), *transactionIDToLink)
		if err != nil {
			a.serverError(w, r, err)
			return
		}
	}

	a.Logger.Debug("Webhook completed successfully")
	a.clientResponse(w, r, http.StatusNoContent)
}

// transfer will create a new transfer transaction from a source account to a destination account with an amount
// defined by the transaction triggering the webhook.
func (a *Application) transfer(w http.ResponseWriter, r *http.Request) {
	body, webhookMessage, err := a.parseRequestMessage(r)
	if err != nil {
		a.serverError(w, r, err)
		return
	}

	configValue, err := a.FireflyConfig.FindConfig(firefly.Transfer, webhookMessage)
	if err != nil {
		a.Logger.Debug("No configuration found", "error", err)
		a.clientError(w, r, http.StatusNotFound)
		return
	}
	config, ok := configValue.(firefly.TransferConfig)
	if !ok {
		a.Logger.Error("Invalid configuration type", "config", configValue)
		a.clientError(w, r, http.StatusInternalServerError)
		return
	}
	a.Logger.Debug("Found configuration", "config", config)

	a.Logger.Debug("Verifying signature", "signature", r.Header.Get("Signature"))
	err = webhookMessage.VerifySignature(r.Header.Get("Signature"), string(body), config.Secret)
	if err != nil {
		a.Logger.Error("Failed validating signature", "header", r.Header.Get("Signature"), "error", err)
		a.clientError(w, r, http.StatusBadRequest)
		return
	}

	content, ok := webhookMessage.Content.(firefly.WebhookMessageTransaction)
	if !ok {
		a.Logger.Error("Invalid content type", "content", webhookMessage.Content)
		a.clientError(w, r, http.StatusBadRequest)
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
			a.Logger.Debug("Transactions source id different from configured one", "transaction", t, "config", config)
			a.clientResponse(w, r, http.StatusNoContent)
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
				a.Logger.Error("Invalid transaction amount", "amount", t.Amount)
				a.clientError(w, r, http.StatusBadRequest)
				return
			}
			amount = *config.ModuloAmount - math.Mod(transactionAmount, *config.ModuloAmount)
		}
		if amount <= zeroWithDelta {
			a.Logger.Debug("No need to create new transaction: remainder lesser than zero", "modulo", amount)
			continue
		}
		created, err2 := a.createTransferTransaction(
			&t,
			amount,
			config,
		)
		if err2 != nil {
			a.serverError(w, r, err2)
			return
		}
		transactionIDToLink = &created.Data.ID
	}

	if transactionIDToLink != nil {
		err = a.FireflyClient.LinkTransactions(config.LinkTypeId, strconv.Itoa(content.ID), *transactionIDToLink)
		if err != nil {
			a.serverError(w, r, err)
			return
		}
	}

	a.Logger.Debug("Webhook completed successfully")
	a.clientResponse(w, r, http.StatusNoContent)
}
