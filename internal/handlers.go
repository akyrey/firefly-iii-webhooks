package internal

import "net/http"

func (a *Application) splitTicket(w http.ResponseWriter, r *http.Request) {
	// TODO:
	// 1. Check if the header contains the signature and verify it
	// 2. Check trigger type - should work only on creation
	// 3. Check response type
	// 4. Parse the request body
	// 5. Perform a request to the Firefly III API to retrieve the transaction
	// 6. Check if the account used is the Ticker Restaurant account
	// 7. Calculate the amount to split, using amount / 8 and amount % 8
	// 8. If the module isn't 0, update this transaction setting the amount to the amount / 8 result
	//    and clone the transaction setting the currency to Satispay and the amount to amount % 8 result
}
