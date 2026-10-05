package internal

import (
	"crypto/hmac"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/akyrey/firefly-iii-webhooks/pkg/firefly"
	"github.com/akyrey/firefly-iii-webhooks/pkg/firefly/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"
)

const (
	splitSourceAccountID  = "1"
	splitPayloadForeign   = "26.00"
	splitTransactionPath  = "/api/v1/transactions/27"
	splitTransactionsPath = "/api/v1/transactions"
	splitLinksPath        = "/api/v1/transaction-links"
)

// recordedRequest is a request received by the fake Firefly API.
type recordedRequest struct {
	Method string
	Path   string
	Body   string
}

// splitFirefly is a fake Firefly API answering the calls made by the split ticket webhook.
type splitFirefly struct {
	mu       sync.Mutex
	requests []recordedRequest
}

func (f *splitFirefly) record(r *http.Request) recordedRequest {
	body, _ := io.ReadAll(r.Body)
	req := recordedRequest{Method: r.Method, Path: r.URL.Path, Body: string(body)}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.requests = append(f.requests, req)
	return req
}

func (f *splitFirefly) writes() []recordedRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	var writes []recordedRequest
	for _, req := range f.requests {
		if req.Method != http.MethodGet {
			writes = append(writes, req)
		}
	}
	return writes
}

// newSplitFirefly starts a fake Firefly API whose GET of the triggering transaction returns the given
// current state (status and transaction json).
func newSplitFirefly(t *testing.T, currentStatus int, currentTransaction string) (*httptest.Server, *splitFirefly) {
	t.Helper()
	fake := &splitFirefly{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := fake.record(r)
		switch {
		case req.Method == http.MethodGet && req.Path == splitTransactionPath:
			w.WriteHeader(currentStatus)
			if currentStatus != http.StatusOK {
				_, _ = w.Write([]byte(`{"message":"Unauthenticated."}`))
				return
			}
			_, _ = fmt.Fprintf(w, `{"data":{"id":"%d","attributes":{"transactions":[%s]}}}`, testGroupID, currentTransaction)
		case req.Method == http.MethodPut && req.Path == splitTransactionPath:
			_, _ = fmt.Fprintf(w,
				`{"data":{"id":"%d","attributes":{"transactions":[{"transaction_journal_id":%q}]}}}`,
				testGroupID, testOriginalJournalID,
			)
		case req.Method == http.MethodPost && req.Path == splitTransactionsPath:
			_, _ = fmt.Fprintf(w,
				`{"data":{"id":%q,"attributes":{"transactions":[{"transaction_journal_id":%q}]}}}`,
				testCreatedGroupID, testCreatedJournalID,
			)
		case req.Method == http.MethodPost && req.Path == splitLinksPath:
			_, _ = w.Write([]byte(`{}`))
		default:
			t.Errorf("unexpected request %s %s", req.Method, req.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, fake
}

func currentSplitTransaction(foreignAmount string, tags ...string) string {
	encodedTags, _ := json.Marshal(append([]string{}, tags...))
	return fmt.Sprintf(`{"user":"1","transaction_journal_id":%q,"foreign_amount":%q,"tags":%s}`,
		testOriginalJournalID, foreignAmount, encodedTags)
}

func splitTicketApplication(baseUrl string) *Application {
	return newTestApplication(baseUrl, firefly.Config{
		firefly.SplitTicket: []firefly.ConfigValue{firefly.SplitTicketConfig{
			Trigger:                          firefly.STORE_TRANSACTION,
			Response:                         firefly.RESPONSE_TRANSACTIONS,
			Secret:                           testSecret,
			Type:                             firefly.WITHDRAWAL,
			SourceAccountId:                  splitSourceAccountID,
			DestinationAccountId:             "4",
			DestinationCurrencyId:            "1",
			DestinationCurrencyDecimalPlaces: 2,
			SplitAmount:                      8,
			LinkTypeId:                       testLinkTypeID,
		}},
	})
}

// signedSplitTicketRequest builds a signed webhook for a ticket withdrawal of 3 tickets worth 26.00 EUR,
// shaped like the messages Firefly III v6.7 sends.
func signedSplitTicketRequest(t *testing.T, trigger firefly.WebhookTrigger) *http.Request {
	t.Helper()
	body := fmt.Sprintf(
		`{"uuid":"23a7fd16-3a55-4cef-85ae-059c666520b7","user_id":1,"user_group_id":1,"trigger":%q,`+
			`"response":"TRANSACTIONS","url":"http://firefly_webhooks:4000/api/v1/webhook/split-ticket","version":"v0",`+
			`"content":{"id":%d,"user":1,"group_title":null,"transactions":[{"user":1,"transaction_journal_id":%q,`+
			`"type":"withdrawal","date":"2026-10-01T12:00:00+02:00","order":0,"currency_id":"26","currency_code":"TKT",`+
			`"currency_symbol":"@","currency_decimal_places":0,"foreign_currency_id":"1","foreign_currency_code":"EUR",`+
			`"foreign_currency_symbol":"€","foreign_currency_decimal_places":2,"amount":"3","foreign_amount":%q,`+
			`"description":"Lunch","source_id":%q,"source_name":"Ticket Restaurant","source_type":"Asset account",`+
			`"destination_id":"6","destination_name":"Restaurant","destination_type":"Expense account",`+
			`"budget_id":"","budget_name":null,"category_id":"","category_name":null,"bill_id":"","bill_name":null,`+
			`"reconciled":false,"notes":null,"tags":[],"longitude":null,"latitude":null,"zoom_level":null}]}}`,
		trigger, testGroupID, testOriginalJournalID, splitPayloadForeign, splitSourceAccountID,
	)
	timestamp := "1610738765"
	mac := hmac.New(sha3.New256, []byte(testSecret))
	_, err := mac.Write([]byte(timestamp + "." + body))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhook/split-ticket", strings.NewReader(body))
	req.Header.Set("Signature", fmt.Sprintf("t=%s,v1=%x", timestamp, mac.Sum(nil)))
	return req
}

func TestSplitTicketSplitsUnprocessedTransaction(t *testing.T) {
	// Arrange
	srv, fake := newSplitFirefly(t, http.StatusOK, currentSplitTransaction(splitPayloadForeign))
	app := splitTicketApplication(srv.URL)
	rec := httptest.NewRecorder()

	// Act
	app.splitTicket(rec, signedSplitTicketRequest(t, firefly.STORE_TRANSACTION))

	// Assert
	assert.Equal(t, http.StatusNoContent, rec.Code)
	writes := fake.writes()
	require.Len(t, writes, 3)

	var update models.UpdateTransactionRequest
	require.Equal(t, http.MethodPut, writes[0].Method)
	require.NoError(t, json.Unmarshal([]byte(writes[0].Body), &update))
	require.Len(t, update.Transactions, 1)
	assert.Equal(t, "3", update.Transactions[0].Amount)
	assert.Equal(t, "24.00", *update.Transactions[0].ForeignAmount)
	assert.Contains(t, update.Transactions[0].Tags, "Webhook: split_ticket")

	var created models.StoreTransactionRequest
	require.Equal(t, splitTransactionsPath, writes[1].Path)
	require.NoError(t, json.Unmarshal([]byte(writes[1].Body), &created))
	require.Len(t, created.Transactions, 1)
	assert.Equal(t, "2.00", created.Transactions[0].Amount)

	var link models.StoreLinkRequest
	require.Equal(t, splitLinksPath, writes[2].Path)
	require.NoError(t, json.Unmarshal([]byte(writes[2].Body), &link))
	assert.Equal(t, models.StoreLinkRequest{
		LinkTypeID: testLinkTypeID,
		InwardID:   testOriginalJournalID,
		OutwardID:  testCreatedJournalID,
	}, link)
}

func TestSplitTicketSkipsTransactionAlreadySplit(t *testing.T) {
	// Arrange: Firefly replays a stored message for a transaction this webhook already updated.
	srv, fake := newSplitFirefly(t, http.StatusOK, currentSplitTransaction("24.00", "Webhook: split_ticket"))
	app := splitTicketApplication(srv.URL)
	rec := httptest.NewRecorder()

	// Act
	app.splitTicket(rec, signedSplitTicketRequest(t, firefly.STORE_TRANSACTION))

	// Assert
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, fake.writes())
}

func TestSplitTicketSkipsStalePayload(t *testing.T) {
	// Arrange: the user edited the foreign amount after Firefly stored the message.
	srv, fake := newSplitFirefly(t, http.StatusOK, currentSplitTransaction("8.00"))
	app := splitTicketApplication(srv.URL)
	rec := httptest.NewRecorder()

	// Act
	app.splitTicket(rec, signedSplitTicketRequest(t, firefly.STORE_TRANSACTION))

	// Assert
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, fake.writes())
}

func TestSplitTicketAcknowledgesMessagesWithoutMatchingConfig(t *testing.T) {
	// Arrange: a webhook with several triggers also delivers updates, which are not configured.
	srv, fake := newSplitFirefly(t, http.StatusOK, currentSplitTransaction(splitPayloadForeign))
	app := splitTicketApplication(srv.URL)
	rec := httptest.NewRecorder()

	// Act
	app.splitTicket(rec, signedSplitTicketRequest(t, firefly.UPDATE_TRANSACTION))

	// Assert
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, fake.requests)
}

func TestSplitTicketFailsWithoutChangesWhenCurrentStateIsUnavailable(t *testing.T) {
	// Arrange: an expired API token makes every Firefly call fail.
	srv, fake := newSplitFirefly(t, http.StatusUnauthorized, "")
	app := splitTicketApplication(srv.URL)
	rec := httptest.NewRecorder()

	// Act
	app.splitTicket(rec, signedSplitTicketRequest(t, firefly.STORE_TRANSACTION))

	// Assert
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Empty(t, fake.writes())
}
