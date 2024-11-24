package internal

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *Application) Routes(config Config) http.Handler {
	mux := http.NewServeMux()

	protected := alice.New(
		// TODO: add authentication middleware
		app.recoverPanic,
		app.logRequest,
		app.secureHeaders,
		app.contentTypeHeader,
	)

	mux.Handle("/api/v1/webhook/split-ticket", protected.ThenFunc(app.splitTicket))
	mux.Handle("/api/v1/webhook/cashback", protected.ThenFunc(app.cashback))
	mux.Handle("/api/v1/webhook/transfer", protected.ThenFunc(app.transfer))

	return protected.Then(mux)
}
