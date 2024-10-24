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
	)

	mux.Handle("/api/v1/webhook/split-ticket", protected.ThenFunc(app.splitTicket))

	return protected.Then(mux)
}
