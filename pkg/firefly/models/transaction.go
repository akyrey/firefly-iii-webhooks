package models

import "time"

type Transaction struct {
	Date                         time.Time   `json:"date"`
	ZoomLevel                    interface{} `json:"zoom_level,omitempty"`
	Longitude                    interface{} `json:"longitude,omitempty"`
	Latitude                     interface{} `json:"latitude,omitempty"`
	InternalReference            interface{} `json:"internal_reference,omitempty"`
	InterestDate                 *time.Time  `json:"interest_date,omitempty"`
	ExternalID                   *int        `json:"external_id,omitempty"`
	CategoryID                   *int        `json:"category_id"`
	CategoryName                 *string     `json:"category_name"`
	BillID                       *int        `json:"bill_id,omitempty"`
	BillName                     *string     `json:"bill_name,omitempty"`
	BookDate                     *time.Time  `json:"book_date,omitempty"`
	SourceIban                   *string     `json:"source_iban,omitempty"`
	BudgetName                   *string     `json:"budget_name,omitempty"`
	SepaEp                       *string     `json:"sepa_ep,omitempty"`
	DestinationIban              *string     `json:"destination_iban,omitempty"`
	SepaDb                       *string     `json:"sepa_db,omitempty"`
	BunqPaymentID                *int        `json:"bunq_payment_id,omitempty"`
	SepaCtOp                     *string     `json:"sepa_ct_op,omitempty"`
	DueDate                      *time.Time  `json:"due_date,omitempty"`
	ProcessDate                  *time.Time  `json:"process_date,omitempty"`
	ForeignAmount                *string     `json:"foreign_amount,omitempty"`
	ForeignCurrencyCode          *string     `json:"foreign_currency_code,omitempty"`
	InvoiceDate                  *time.Time  `json:"invoice_date,omitempty"`
	ForeignCurrencyID            *int        `json:"foreign_currency_id,omitempty"`
	ForeignCurrencySymbol        *string     `json:"foreign_currency_symbol,omitempty"`
	SepaCtID                     *int        `json:"sepa_ct_id,omitempty"`
	SepaCountry                  *string     `json:"sepa_country,omitempty"`
	BudgetID                     *int        `json:"budget_id,omitempty"`
	ForeignCurrencyDecimalPlaces *int        `json:"foreign_currency_decimal_places,omitempty"`
	SepaCi                       *string     `json:"sepa_ci,omitempty"`
	SepaCc                       *string     `json:"sepa_cc,omitempty"`
	Notes                        *string     `json:"notes,omitempty"`
	SepaBatchID                  *int        `json:"sepa_batch_id,omitempty"`
	RecurrenceID                 *int        `json:"recurrence_id,omitempty"`
	PaymentDate                  *time.Time  `json:"payment_date,omitempty"`
	DestinationName              string      `json:"destination_name"`
	SourceName                   string      `json:"source_name"`
	OriginalSource               string      `json:"original_source"`
	CurrencyCode                 string      `json:"currency_code"`
	Type                         string      `json:"type"`
	Amount                       string      `json:"amount"`
	DestinationType              string      `json:"destination_type"`
	SourceType                   string      `json:"source_type"`
	Description                  string      `json:"description"`
	CurrencySymbol               string      `json:"currency_symbol"`
	Tags                         []string    `json:"tags"`
	CurrencyID                   int         `json:"currency_id"`
	SourceID                     int         `json:"source_id"`
	DestinationID                int         `json:"destination_id"`
	TransactionJournalID         int         `json:"transaction_journal_id"`
	User                         int         `json:"user"`
	CurrencyDecimalPlaces        int         `json:"currency_decimal_places"`
	Order                        int         `json:"order"`
	Reconciled                   bool        `json:"reconciled"`
}
