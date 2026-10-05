package firefly

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTransactionReturnsCurrentTransactionGroup(t *testing.T) {
	// Arrange
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/api/v1/transactions/27", r.URL.Path)
		assert.Equal(t, "Bearer key", r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(`{"data":{"id":"27","attributes":{"user":"1","transactions":[{` +
			`"user":"1","transaction_journal_id":"42","foreign_amount":"26.00","tags":["Webhook: split_ticket"]}]}}}`))
	}))
	t.Cleanup(srv.Close)
	client := NewFirefly(srv.URL, WithApiKey("key"))

	// Act
	group, err := client.GetTransaction(27)

	// Assert
	require.NoError(t, err)
	require.Len(t, group.Data.Attributes.Transactions, 1)
	transaction := group.Data.Attributes.Transactions[0]
	assert.Equal(t, "42", transaction.TransactionJournalID)
	assert.Equal(t, "26.00", transaction.ForeignAmount)
	assert.Equal(t, []string{"Webhook: split_ticket"}, transaction.Tags)
}

func TestGetTransactionReturnsErrorOnUnauthorized(t *testing.T) {
	// Arrange
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Unauthenticated."}`))
	}))
	t.Cleanup(srv.Close)
	client := NewFirefly(srv.URL, WithApiKey("expired"))

	// Act
	group, err := client.GetTransaction(27)

	// Assert
	require.Error(t, err)
	assert.Nil(t, group)
}
