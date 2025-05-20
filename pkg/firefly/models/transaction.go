package models

import (
	"time"
)

type Transaction struct {
	Date                         time.Time  `json:"date"`
	ZoomLevel                    any        `json:"zoom_level,omitempty"`
	Longitude                    any        `json:"longitude,omitempty"`
	Latitude                     any        `json:"latitude,omitempty"`
	InternalReference            any        `json:"internal_reference,omitempty"`
	InterestDate                 *time.Time `json:"interest_date,omitempty"`
	ExternalID                   *int       `json:"external_id,omitempty"`
	CategoryID                   *int       `json:"category_id"`
	CategoryName                 *string    `json:"category_name"`
	BillID                       *string    `json:"bill_id,omitempty"`
	BillName                     *string    `json:"bill_name,omitempty"`
	BookDate                     *time.Time `json:"book_date,omitempty"`
	SourceIban                   *string    `json:"source_iban,omitempty"`
	BudgetName                   *string    `json:"budget_name,omitempty"`
	SepaEp                       *string    `json:"sepa_ep,omitempty"`
	DestinationIban              *string    `json:"destination_iban,omitempty"`
	SepaDb                       *string    `json:"sepa_db,omitempty"`
	BunqPaymentID                *int       `json:"bunq_payment_id,omitempty"`
	SepaCtOp                     *string    `json:"sepa_ct_op,omitempty"`
	DueDate                      *time.Time `json:"due_date,omitempty"`
	ProcessDate                  *time.Time `json:"process_date,omitempty"`
	ForeignAmount                *string    `json:"foreign_amount,omitempty"`
	ForeignCurrencyCode          *string    `json:"foreign_currency_code,omitempty"`
	InvoiceDate                  *time.Time `json:"invoice_date,omitempty"`
	ForeignCurrencyID            *string    `json:"foreign_currency_id,omitempty"`
	ForeignCurrencySymbol        *string    `json:"foreign_currency_symbol,omitempty"`
	SepaCtID                     *int       `json:"sepa_ct_id,omitempty"`
	SepaCountry                  *string    `json:"sepa_country,omitempty"`
	BudgetID                     *string    `json:"budget_id,omitempty"`
	ForeignCurrencyDecimalPlaces *int       `json:"foreign_currency_decimal_places,omitempty"`
	SepaCi                       *string    `json:"sepa_ci,omitempty"`
	SepaCc                       *string    `json:"sepa_cc,omitempty"`
	Notes                        *string    `json:"notes,omitempty"`
	SepaBatchID                  *int       `json:"sepa_batch_id,omitempty"`
	RecurrenceID                 *int       `json:"recurrence_id,omitempty"`
	PaymentDate                  *time.Time `json:"payment_date,omitempty"`
	DestinationName              string     `json:"destination_name"`
	SourceName                   string     `json:"source_name"`
	OriginalSource               string     `json:"original_source"`
	CurrencyCode                 string     `json:"currency_code"`
	Type                         string     `json:"type"`
	Amount                       string     `json:"amount"`
	DestinationType              string     `json:"destination_type"`
	SourceType                   string     `json:"source_type"`
	Description                  string     `json:"description"`
	CurrencySymbol               string     `json:"currency_symbol"`
	Tags                         []string   `json:"tags"`
	CurrencyID                   string     `json:"currency_id"`
	SourceID                     string     `json:"source_id"`
	DestinationID                string     `json:"destination_id"`
	TransactionJournalID         string     `json:"transaction_journal_id,omitempty"`
	User                         int        `json:"user"`
	CurrencyDecimalPlaces        int        `json:"currency_decimal_places"`
	Order                        int        `json:"order"`
	Reconciled                   bool       `json:"reconciled"`
}

type TransactionResponse struct {
	User                         string    `json:"user"`
	TransactionJournalID         string    `json:"transaction_journal_id"`
	Type                         string    `json:"type"`
	Date                         time.Time `json:"date"`
	Order                        int       `json:"order"`
	CurrencyID                   string    `json:"currency_id"`
	CurrencyCode                 string    `json:"currency_code"`
	CurrencySymbol               string    `json:"currency_symbol"`
	CurrencyName                 string    `json:"currency_name"`
	CurrencyDecimalPlaces        int       `json:"currency_decimal_places"`
	ForeignCurrencyID            string    `json:"foreign_currency_id"`
	ForeignCurrencyCode          string    `json:"foreign_currency_code"`
	ForeignCurrencySymbol        string    `json:"foreign_currency_symbol"`
	ForeignCurrencyDecimalPlaces int       `json:"foreign_currency_decimal_places"`
	Amount                       string    `json:"amount"`
	ForeignAmount                string    `json:"foreign_amount"`
	Description                  string    `json:"description"`
	SourceID                     string    `json:"source_id"`
	SourceName                   string    `json:"source_name"`
	SourceIban                   string    `json:"source_iban"`
	SourceType                   string    `json:"source_type"`
	DestinationID                string    `json:"destination_id"`
	DestinationName              string    `json:"destination_name"`
	DestinationIban              string    `json:"destination_iban"`
	DestinationType              string    `json:"destination_type"`
	BudgetID                     string    `json:"budget_id"`
	BudgetName                   string    `json:"budget_name"`
	CategoryID                   string    `json:"category_id"`
	CategoryName                 string    `json:"category_name"`
	BillID                       string    `json:"bill_id"`
	BillName                     string    `json:"bill_name"`
	Reconciled                   bool      `json:"reconciled"`
	Notes                        string    `json:"notes"`
	Tags                         any       `json:"tags"`
	InternalReference            string    `json:"internal_reference"`
	ExternalID                   string    `json:"external_id"`
	ExternalURL                  string    `json:"external_url"`
	OriginalSource               string    `json:"original_source"`
	RecurrenceID                 string    `json:"recurrence_id"`
	RecurrenceTotal              int       `json:"recurrence_total"`
	RecurrenceCount              int       `json:"recurrence_count"`
	BunqPaymentID                string    `json:"bunq_payment_id"`
	ImportHashV2                 string    `json:"import_hash_v2"`
	SepaCc                       string    `json:"sepa_cc"`
	SepaCtOp                     string    `json:"sepa_ct_op"`
	SepaCtID                     string    `json:"sepa_ct_id"`
	SepaDb                       string    `json:"sepa_db"`
	SepaCountry                  string    `json:"sepa_country"`
	SepaEp                       string    `json:"sepa_ep"`
	SepaCi                       string    `json:"sepa_ci"`
	SepaBatchID                  string    `json:"sepa_batch_id"`
	InterestDate                 time.Time `json:"interest_date"`
	BookDate                     time.Time `json:"book_date"`
	ProcessDate                  time.Time `json:"process_date"`
	DueDate                      time.Time `json:"due_date"`
	PaymentDate                  time.Time `json:"payment_date"`
	InvoiceDate                  time.Time `json:"invoice_date"`
	Latitude                     float64   `json:"latitude"`
	Longitude                    float64   `json:"longitude"`
	ZoomLevel                    int       `json:"zoom_level"`
	HasAttachments               bool      `json:"has_attachments"`
}

type StoreTransactionRequest struct {
	GroupTitle           string        `json:"group_title"`
	Transactions         []Transaction `json:"transactions"`
	ErrorIfDuplicateHash bool          `json:"error_if_duplicate_hash"`
	ApplyRules           bool          `json:"apply_rules"`
	FireWebhooks         bool          `json:"fire_webhooks"`
}

type UpdateTransactionRequest struct {
	GroupTitle   string        `json:"group_title"`
	Transactions []Transaction `json:"transactions"`
	ApplyRules   bool          `json:"apply_rules"`
	FireWebhooks bool          `json:"fire_webhooks"`
}

type UpsertTransactionResponse struct {
	Data struct {
		Type       string `json:"type"`
		ID         string `json:"id"`
		Attributes struct {
			CreatedAt    string                `json:"created_at"`
			UpdateAt     string                `json:"updated_at"`
			User         string                `json:"user"`
			GroupTitle   string                `json:"group_title"`
			Transactions []TransactionResponse `json:"transactions"`
		} `json:"attributes"`
	} `json:"data"`
}
