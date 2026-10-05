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
	"time"

	"github.com/akyrey/firefly-iii-webhooks/pkg/firefly"
	"github.com/akyrey/firefly-iii-webhooks/pkg/firefly/models"
	"github.com/akyrey/firefly-iii-webhooks/pkg/utils"
	"github.com/jinzhu/copier"
)

// parseRequestMessage will parse the request message and return the body and the webhook message.
func (a *Application) parseRequestMessage(r *http.Request) (body []byte, webhookMessage firefly.WebhookMessage, err error) {
	body, err = io.ReadAll(r.Body)
	if err != nil {
		return nil, firefly.WebhookMessage{}, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(r.Body)
	err = json.Unmarshal(body, &webhookMessage)
	if err != nil {
		return nil, firefly.WebhookMessage{}, err
	}
	a.Logger.Debug("Received body", "body", webhookMessage)
	return body, webhookMessage, nil
}

// webhookTag returns the tag attached to every transaction touched by the given webhook action.
func webhookTag(action firefly.ConfigType) string {
	return fmt.Sprintf("%s %s", firefly.WEBHOOK_TAG_PREFIX, action)
}

// splitTicketAlreadyHandled fetches the current state of the transaction group and reports whether the split
// must be skipped: because this webhook already split it, or because it was edited after Firefly stored the
// message (its foreign amount no longer matches the payload one).
func (a *Application) splitTicketAlreadyHandled(groupID int, payloadForeignAmount float64, decimalPlaces int) (bool, error) {
	current, err := a.FireflyClient.GetTransaction(groupID)
	if err != nil {
		return false, fmt.Errorf("failed fetching current state of transaction group %d: %w", groupID, err)
	}

	transactions := current.Data.Attributes.Transactions
	if len(transactions) != 1 {
		a.Logger.Info("Transaction group no longer has exactly one transaction, skipping split", "groupID", groupID, "count", len(transactions))
		return true, nil
	}
	t := transactions[0]
	if slices.Contains(t.Tags, webhookTag(firefly.SplitTicket)) {
		a.Logger.Info("Transaction already split, skipping replayed message", "groupID", groupID)
		return true, nil
	}

	currentForeignAmount, err := strconv.ParseFloat(strings.TrimSpace(t.ForeignAmount), 64)
	if err != nil || math.Abs(currentForeignAmount-payloadForeignAmount) >= math.Pow10(-decimalPlaces) {
		a.Logger.Info("Transaction changed since the message was stored, skipping split",
			"groupID", groupID, "current foreign amount", t.ForeignAmount, "payload foreign amount", payloadForeignAmount)
		return true, nil
	}

	return false, nil
}

// updateSplitTransaction will update the transaction amount to the number of tickets and the foreign amount
// to the value they cover.
func (a *Application) updateSplitTransaction(
	t *models.Transaction,
	contentID int,
	split ticketSplit,
) (*models.UpsertTransactionResponse, error) {
	updatedAmount := fmt.Sprintf("%.[2]*[1]f", float64(split.Tickets), t.CurrencyDecimalPlaces)
	updatedForeignAmount := fmt.Sprintf("%.[2]*[1]f", split.Covered, *t.ForeignCurrencyDecimalPlaces)
	var tToUpdate models.Transaction
	err := copier.Copy(&tToUpdate, t)
	if err != nil {
		a.Logger.Error("Failed copying transaction", "error", err)
		return nil, err
	}

	tToUpdate.Amount = updatedAmount
	tToUpdate.ForeignAmount = &updatedForeignAmount
	tToUpdate.Tags = append(tToUpdate.Tags, webhookTag(firefly.SplitTicket))
	tToUpdate.TransactionJournalID = ""
	a.Logger.Debug("Updating transaction amount, foreign amount and tags", "contentID", contentID, "transaction", tToUpdate)
	return a.FireflyClient.UpdateTransaction(
		contentID,
		&models.UpdateTransactionRequest{
			ApplyRules:   true,
			FireWebhooks: true,
			Transactions: []models.Transaction{tToUpdate},
		})
}

// createSplitTransaction will create a new transaction with the remaining amount.
func (a *Application) createSplitTransaction(
	t *models.Transaction,
	remainder float64,
	currencyDecimalPlaces int,
	accountID string,
	currencyID string,
) (*models.UpsertTransactionResponse, error) {
	remainderAmount := fmt.Sprintf("%.[2]*[1]f", remainder, currencyDecimalPlaces)
	tToCreate := models.Transaction{
		Amount:        remainderAmount,
		SourceID:      accountID,
		CurrencyID:    currencyID,
		DestinationID: t.DestinationID,
		User:          t.User,
		Type:          string(firefly.WITHDRAWAL),
		Description:   t.Description,
		BudgetID:      t.BudgetID,
		CategoryID:    t.CategoryID,
		Tags:          append(t.Tags, webhookTag(firefly.SplitTicket)),
		Date:          t.Date.Add(time.Second),
		Notes:         t.Notes,
	}
	a.Logger.Debug("Creating transaction", "transaction", tToCreate)
	return a.FireflyClient.CreateTransaction(&models.StoreTransactionRequest{
		ApplyRules:           true,
		ErrorIfDuplicateHash: true,
		FireWebhooks:         true,
		Transactions:         []models.Transaction{tToCreate},
	})
}

// createCashbackTransaction will create a new transaction with the cashback amount.
func (a *Application) createCashbackTransaction(
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
	a.Logger.Debug("Creating transaction", "transaction", tToCreate)
	return a.FireflyClient.CreateTransaction(&models.StoreTransactionRequest{
		ApplyRules:           true,
		ErrorIfDuplicateHash: true,
		FireWebhooks:         true,
		Transactions:         []models.Transaction{tToCreate},
	})
}

// linkCreatedTransaction will link the transaction that triggered the webhook with the one created from it.
// Firefly links transaction journals, so both ids must be transaction journal ids and not transaction group ids.
func (a *Application) linkCreatedTransaction(
	linkTypeID string,
	original *models.Transaction,
	created *models.UpsertTransactionResponse,
) error {
	if len(created.Data.Attributes.Transactions) != 1 {
		a.Logger.Debug("Created transaction doesn't have exactly one transaction, skipping linking", "created", created)
		return nil
	}
	inwardID := original.TransactionJournalID
	outwardID := created.Data.Attributes.Transactions[0].TransactionJournalID
	if inwardID == "" || outwardID == "" {
		return fmt.Errorf("missing transaction journal id to link: inward %q, outward %q", inwardID, outwardID)
	}

	a.Logger.Debug("Linking transactions", "initial id", inwardID, "created id", outwardID, "link type", linkTypeID)
	if err := a.FireflyClient.LinkTransactions(linkTypeID, inwardID, outwardID); err != nil {
		return fmt.Errorf("failed linking transaction journals %s and %s: %w", inwardID, outwardID, err)
	}
	return nil
}

// createTransferTransaction will create a new transaction with the cashback amount.
func (a *Application) createTransferTransaction(
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
	a.Logger.Debug("Creating transaction", "transaction", tToCreate)
	return a.FireflyClient.CreateTransaction(&models.StoreTransactionRequest{
		ApplyRules:           true,
		ErrorIfDuplicateHash: true,
		FireWebhooks:         true,
		Transactions:         []models.Transaction{tToCreate},
	})
}
