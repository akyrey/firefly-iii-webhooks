package internal

import (
	"crypto/hmac"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/akyrey/firefly-iii-webhooks/pkg/firefly"
	"github.com/akyrey/firefly-iii-webhooks/pkg/firefly/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/sha3"
)

const (
	testSecret            = "secret"
	testGroupID           = 27
	testOriginalJournalID = "42"
	testCreatedGroupID    = "500"
	testCreatedJournalID  = "501"
	testLinkTypeID        = "1"
)

// fakeFirefly records link requests and answers transaction creations with a fixed group and journal id.
func fakeFirefly(t *testing.T) (*httptest.Server, *[]models.StoreLinkRequest) {
	t.Helper()
	links := []models.StoreLinkRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/transactions":
			_, _ = fmt.Fprintf(w,
				`{"data":{"id":%q,"attributes":{"transactions":[{"transaction_journal_id":%q}]}}}`,
				testCreatedGroupID, testCreatedJournalID,
			)
		case "/api/v1/transaction-links":
			var link models.StoreLinkRequest
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(body, &link))
			links = append(links, link)
			_, _ = w.Write([]byte(`{}`))
		default:
			t.Errorf("unexpected request to %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, &links
}

func newTestApplication(baseUrl string, config firefly.Config) *Application {
	return &Application{
		FireflyClient: firefly.NewFirefly(baseUrl, firefly.WithApiKey("key")),
		FireflyConfig: &config,
		Logger:        slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

func signedWebhookRequest(t *testing.T, path, transactionType, sourceID, destinationID, tag string) *http.Request {
	t.Helper()
	body := fmt.Sprintf(
		`{"trigger":"STORE_TRANSACTION","response":"TRANSACTIONS","content":{"id":%d,"transactions":[{`+
			`"transaction_journal_id":%q,"type":%q,"date":"2024-10-31T17:37:00+01:00","amount":"3.40",`+
			`"currency_decimal_places":2,"source_id":%q,"destination_id":%q,"tags":[%q]}]}}`,
		testGroupID, testOriginalJournalID, transactionType, sourceID, destinationID, tag,
	)
	timestamp := "1610738765"
	mac := hmac.New(sha3.New256, []byte(testSecret))
	_, err := mac.Write([]byte(timestamp + "." + body))
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Signature", fmt.Sprintf("t=%s,v1=%x", timestamp, mac.Sum(nil)))
	return req
}

func TestCashbackLinksTransactionJournalIDs(t *testing.T) {
	// Arrange
	srv, links := fakeFirefly(t)
	app := newTestApplication(srv.URL, firefly.Config{
		firefly.Cashback: []firefly.ConfigValue{firefly.CashbackConfig{
			Trigger:                          firefly.STORE_TRANSACTION,
			Response:                         firefly.RESPONSE_TRANSACTIONS,
			Secret:                           testSecret,
			Type:                             firefly.WITHDRAWAL,
			SourceAccountId:                  "4",
			SourceMustHaveTag:                "Cashback",
			DepositSourceAccountId:           "10",
			DestinationAccountId:             "4",
			DestinationCurrencyId:            "1",
			DestinationCurrencyDecimalPlaces: 2,
			Amount:                           0.02,
			LinkTypeId:                       testLinkTypeID,
		}},
	})
	req := signedWebhookRequest(t, "/api/v1/webhook/cashback", "withdrawal", "4", "6", "Cashback")
	rec := httptest.NewRecorder()

	// Act
	app.cashback(rec, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, []models.StoreLinkRequest{{
		LinkTypeID: testLinkTypeID,
		InwardID:   testOriginalJournalID,
		OutwardID:  testCreatedJournalID,
	}}, *links)
}

func TestTransferLinksTransactionJournalIDs(t *testing.T) {
	// Arrange
	srv, links := fakeFirefly(t)
	moduloAmount := 1.0
	app := newTestApplication(srv.URL, firefly.Config{
		firefly.Transfer: []firefly.ConfigValue{firefly.TransferConfig{
			Trigger:                          firefly.STORE_TRANSACTION,
			Response:                         firefly.RESPONSE_TRANSACTIONS,
			Secret:                           testSecret,
			Type:                             firefly.WITHDRAWAL,
			SourceAccountId:                  "4",
			SourceMustHaveTag:                "Satispay Money Box",
			DestinationAccountId:             "8",
			DestinationCurrencyId:            "1",
			DestinationCurrencyDecimalPlaces: 2,
			ModuloAmount:                     &moduloAmount,
			LinkTypeId:                       testLinkTypeID,
		}},
	})
	req := signedWebhookRequest(t, "/api/v1/webhook/transfer", "withdrawal", "4", "6", "Satispay Money Box")
	rec := httptest.NewRecorder()

	// Act
	app.transfer(rec, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, []models.StoreLinkRequest{{
		LinkTypeID: testLinkTypeID,
		InwardID:   testOriginalJournalID,
		OutwardID:  testCreatedJournalID,
	}}, *links)
}

func TestHandlersAcknowledgeMessagesWithoutMatchingConfig(t *testing.T) {
	// A webhook with several triggers also delivers messages no config handles: they must be acknowledged,
	// otherwise Firefly III keeps re-sending them.
	tests := []struct {
		name    string
		path    string
		handler func(*Application) http.HandlerFunc
		config  firefly.Config
	}{
		{
			name:    "cashback",
			path:    "/api/v1/webhook/cashback",
			handler: func(a *Application) http.HandlerFunc { return a.cashback },
			config: firefly.Config{firefly.Cashback: []firefly.ConfigValue{firefly.CashbackConfig{
				Trigger:  firefly.UPDATE_TRANSACTION,
				Response: firefly.RESPONSE_TRANSACTIONS,
				Secret:   testSecret,
				Type:     firefly.WITHDRAWAL,
			}}},
		},
		{
			name:    "transfer",
			path:    "/api/v1/webhook/transfer",
			handler: func(a *Application) http.HandlerFunc { return a.transfer },
			config: firefly.Config{firefly.Transfer: []firefly.ConfigValue{firefly.TransferConfig{
				Trigger:  firefly.UPDATE_TRANSACTION,
				Response: firefly.RESPONSE_TRANSACTIONS,
				Secret:   testSecret,
				Type:     firefly.WITHDRAWAL,
			}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			srv, links := fakeFirefly(t)
			app := newTestApplication(srv.URL, tt.config)
			req := signedWebhookRequest(t, tt.path, "withdrawal", "4", "6", "Cashback")
			rec := httptest.NewRecorder()

			// Act
			tt.handler(app)(rec, req)

			// Assert
			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Empty(t, *links)
		})
	}
}
