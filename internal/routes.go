package internal

import (
	"net/http"

	"github.com/justinas/alice"
)

func (a *Application) Routes(config Config) http.Handler {
	mux := http.NewServeMux()

	protected := alice.New(
		// TODO: add authentication middleware
		a.recoverPanic,
		a.logRequest,
		a.secureHeaders,
		a.contentTypeHeader,
	)

	mux.Handle("/api/v1/webhook/split-ticket", protected.ThenFunc(a.splitTicket))
	mux.Handle("/api/v1/webhook/cashback", protected.ThenFunc(a.cashback))
	mux.Handle("/api/v1/webhook/transfer", protected.ThenFunc(a.transfer))

	return protected.Then(mux)
}
