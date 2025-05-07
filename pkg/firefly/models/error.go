package models

import "fmt"

// FireflyErrReply is the error response body for firefly client.
type FireflyErrReply struct {
	Code      int                 `json:"code"`
	Status    string              `json:"status"`
	Message   string              `json:"message"`
	Exception string              `json:"exception,omitempty"`
	Errors    map[string][]string `json:"errors,omitempty"`
}

func (e FireflyErrReply) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("Error response from Firefly API (status %d - %s): %s", e.Code, e.Status, e.Message)
	}

	return fmt.Sprintf("Unknown error (status %d - %s)", e.Code, e.Status)
}
