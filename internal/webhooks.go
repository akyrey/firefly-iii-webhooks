package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/akyrey/firefly-iii-webhooks/pkg/firefly"
	"github.com/akyrey/firefly-iii-webhooks/pkg/firefly/models"
	"github.com/akyrey/firefly-iii-webhooks/pkg/utils"
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
func (app *Application) updateSplitTransaction(
	t *models.Transaction,
	contentID int,
	division float64,
	splitAmount float64,
) (*models.UpsertTransactionResponse, error) {
	updatedForeignAmountF := division * splitAmount
	updatedAmount := fmt.Sprintf("%.[2]*[1]f", division, t.CurrencyDecimalPlaces)
	updatedForeignAmount := fmt.Sprintf("%.[2]*[1]f", updatedForeignAmountF, *t.ForeignCurrencyDecimalPlaces)
	var tToUpdate models.Transaction
	err := copier.Copy(&tToUpdate, t)
	if err != nil {
		app.Logger.Error("Failed copying transaction", "error", err)
		return nil, err
	}

	tToUpdate.Amount = updatedAmount
	tToUpdate.ForeignAmount = &updatedForeignAmount
	tToUpdate.Tags = append(tToUpdate.Tags, fmt.Sprintf("%s %s", firefly.WEBHOOK_TAG_PREFIX, firefly.SplitTicket))
	tToUpdate.TransactionJournalID = 0
	app.Logger.Debug("Updating transaction amount, foreign amount and tags", "contentID", contentID, "transaction", tToUpdate)
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
	modulo float64,
	currencyDecimalPlaces int,
	accountID int,
	currencyID int,
) (*models.UpsertTransactionResponse, error) {
	moduloAmount := fmt.Sprintf("%.[2]*[1]f", modulo, currencyDecimalPlaces)
	tToCreate := models.Transaction{
		Amount:        moduloAmount,
		SourceID:      accountID,
		CurrencyID:    currencyID,
		DestinationID: t.DestinationID,
		User:          t.User,
		Type:          string(firefly.WITHDRAWAL),
		Description:   t.Description,
		BudgetID:      t.BudgetID,
		CategoryID:    t.CategoryID,
		Tags:          append(t.Tags, fmt.Sprintf("%s %s", firefly.WEBHOOK_TAG_PREFIX, firefly.SplitTicket)),
		Date:          t.Date,
		Notes:         t.Notes,
	}
	app.Logger.Debug("Creating transaction", "transaction", tToCreate)
	return app.FireflyClient.CreateTransaction(&models.StoreTransactionRequest{
		ApplyRules:           true,
		ErrorIfDuplicateHash: true,
		FireWebhooks:         true,
		Transactions:         []models.Transaction{tToCreate},
	})
}

// createCashbackTransaction will create a new transaction with the cashback amount.
func (app *Application) createCashbackTransaction(
	t *models.Transaction,
	config firefly.CashbackConfig,
) (*models.UpsertTransactionResponse, error) {
	cashbackAmount := fmt.Sprintf("%.[2]*[1]f", config.Amount, config.DestinationCurrencyDecimalPlaces)
	// We need to filter mustHaveTag to avoid creating an infinite loop and previously added webhooks tags.
	tags := utils.Filter(
		t.Tags,
		func(tag string) bool {
			return tag != config.SourceMustHaveTag && !strings.HasPrefix(tag, firefly.WEBHOOK_TAG_PREFIX)
		},
	)
	tags = append(tags, fmt.Sprintf("%s %s", firefly.WEBHOOK_TAG_PREFIX, firefly.Cashback))
	tToCreate := models.Transaction{
		Amount:        cashbackAmount,
		SourceID:      config.DepositSourceAccountId,
		CurrencyID:    config.DestinationCurrencyId,
		DestinationID: config.DestinationAccountId,
		User:          t.User,
		Type:          string(firefly.DEPOSIT),
		Description:   config.Title,
		BudgetID:      t.BudgetID,
		CategoryID:    &config.CategoryID,
		Tags:          tags,
		Date:          t.Date,
		Notes:         t.Notes,
	}
	app.Logger.Debug("Creating transaction", "transaction", tToCreate)
	return app.FireflyClient.CreateTransaction(&models.StoreTransactionRequest{
		ApplyRules:           true,
		ErrorIfDuplicateHash: true,
		FireWebhooks:         true,
		Transactions:         []models.Transaction{tToCreate},
	})
}

// createTransferTransaction will create a new transaction with the cashback amount.
func (app *Application) createTransferTransaction(
	t *models.Transaction,
	amount float64,
	config firefly.TransferConfig,
) (*models.UpsertTransactionResponse, error) {
	transferAmount := fmt.Sprintf("%.[2]*[1]f", amount, config.DestinationCurrencyDecimalPlaces)
	tags := []string{fmt.Sprintf("%s %s", firefly.WEBHOOK_TAG_PREFIX, firefly.Transfer)}
	tToCreate := models.Transaction{
		Amount:        transferAmount,
		SourceID:      config.SourceAccountId,
		CurrencyID:    config.DestinationCurrencyId,
		DestinationID: config.DestinationAccountId,
		User:          t.User,
		Type:          string(firefly.TRANSFER),
		Description:   config.Title,
		BudgetID:      t.BudgetID,
		CategoryID:    &config.CategoryID,
		Tags:          tags,
		Date:          t.Date,
		Notes:         t.Notes,
	}
	app.Logger.Debug("Creating transaction", "transaction", tToCreate)
	return app.FireflyClient.CreateTransaction(&models.StoreTransactionRequest{
		ApplyRules:           true,
		ErrorIfDuplicateHash: true,
		FireWebhooks:         true,
		Transactions:         []models.Transaction{tToCreate},
	})
}
