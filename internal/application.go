package internal

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/akyrey/firefly-iii-webhooks/pkg/firefly"
)

type Application struct {
	Config        Config
	FireflyConfig firefly.Config
	Logger        *slog.Logger
}

func (a Application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	a.Logger.Error(
		err.Error(),
		slog.String("method", r.Method),
		slog.String("uri", r.URL.RequestURI()),
		slog.String("trace", string(debug.Stack())),
	)
	message := a.formatErrorMessage(w, r, err.Error())
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(message))
}

func (a Application) clientError(w http.ResponseWriter, r *http.Request, status int) {
	message := a.formatErrorMessage(w, r, http.StatusText(status))
	w.WriteHeader(status)
	w.Write([]byte(message))
}

func (a Application) notFound(w http.ResponseWriter, r *http.Request) {
	a.clientError(w, r, http.StatusNotFound)
}

// formatErrorMessage will return an error message in the requested format.
func (a Application) formatErrorMessage(w http.ResponseWriter, r *http.Request, message string) string {
	// jsonError, err := json.Marshal(models.ErrorResponse{Message: message})
	// if err == nil {
	// 	return string(jsonError)
	// }

	return message
}

func (a Application) clientResponse(w http.ResponseWriter, r *http.Request, status int, data ...any) {
	w.WriteHeader(status)
	// err := json.NewEncoder(w).Encode(models.DataResponse{Data: data})
	// if err != nil {
	// 	serverError(logger, w, r, err)
	// }
	return
}
