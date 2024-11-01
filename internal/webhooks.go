package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"

	"github.com/akyrey/firefly-iii-webhooks/pkg/firefly"
	"github.com/akyrey/firefly-iii-webhooks/pkg/firefly/models"
	"github.com/jinzhu/copier"
)

// parseRequestMessage will parse the request message and return the body and the webhook message.
func (app *Application) parseRequestMessage(r *http.Request) (body []byte, webhookMessage firefly.WebhookMessage, err error) {
	body, err = io.ReadAll(r.Body)
	if err != nil {
		return nil, firefly.WebhookMessage{}, err
	}
	defer r.Body.Close()
	err = json.Unmarshal(body, &webhookMessage)
	if err != nil {
		return nil, firefly.WebhookMessage{}, err
	}
	app.Logger.Debug("Received body", "body", webhookMessage)
	return body, webhookMessage, nil
}

// updateSplitTransaction will update the transaction with the new amount and foreign amount.
func (app *Application) updateSplitTransaction(t *models.Transaction, contentID int, messageUUID string, foreignAmount, splitAmount float64) error {
	division := math.Floor(foreignAmount / splitAmount)
	updatedForeignAmountF := division * splitAmount
	updatedAmount := fmt.Sprintf("%.[2]*[1]f", division, t.CurrencyDecimalPlaces)
	updatedForeignAmount := fmt.Sprintf("%.[2]*[1]f", updatedForeignAmountF, *t.ForeignCurrencyDecimalPlaces)
	var tToUpdate models.Transaction
	err := copier.Copy(&tToUpdate, t)
	if err != nil {
		app.Logger.Error("Failed copying transaction", "error", err)
		return err
	}

	tToUpdate.Amount = updatedAmount
	tToUpdate.ForeignAmount = &updatedForeignAmount
	tToUpdate.Tags = append(tToUpdate.Tags, "Webhook: split_ticket", fmt.Sprintf("Webhook uuid: %s", messageUUID))
	app.Logger.Debug("Updating transaction amount, foreign amount and tags", "transaction", tToUpdate)
	return app.FireflyClient.UpdateTransaction(
		contentID,
		&models.UpdateTransactionRequest{
			ApplyRules:   true,
			FireWebhooks: true,
			Transactions: []models.Transaction{tToUpdate},
		})
}

// createSplitTransaction will create a new transaction with the remaining amount.
func (app *Application) createSplitTransaction(
	t *models.Transaction,
	messageUUID string,
	modulo float64,
	currencyDecimalPlaces int,
	accountID int,
	currencyID int,
) error {
	moduloAmount := fmt.Sprintf("%.[2]*[1]f", modulo, currencyDecimalPlaces)
	var tToCreate models.Transaction
	err := copier.Copy(&tToCreate, &t)
	if err != nil {
		return err
	}
	tToCreate.Amount = moduloAmount
	tToCreate.SourceID = accountID
	tToCreate.CurrencyID = currencyID
	tToCreate.ForeignAmount = nil
	tToCreate.ForeignCurrencyID = nil
	tToCreate.ForeignCurrencyCode = nil
	tToCreate.ForeignCurrencyDecimalPlaces = nil
	tToCreate.ForeignCurrencySymbol = nil
	tToCreate.Tags = append(tToCreate.Tags, "Webhook: split_ticket", fmt.Sprintf("Webhook uuid: %s", messageUUID))
	app.Logger.Debug("Creating transaction", "transaction", tToCreate)
	return app.FireflyClient.CreateTransaction(&models.StoreTransactionRequest{
		ApplyRules:           true,
		ErrorIfDuplicateHash: true,
		FireWebhooks:         true,
		Transactions:         []models.Transaction{tToCreate},
	})
}
