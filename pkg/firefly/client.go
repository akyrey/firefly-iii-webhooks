package firefly

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/akyrey/firefly-iii-webhooks/pkg/assert"
	"github.com/akyrey/firefly-iii-webhooks/pkg/firefly/models"
)

// Firefly client used to interact with the Firefly III API.
type Firefly struct {
	httpClient *http.Client
	baseUrl    string
	// Optional configuration options
	fireflyOpts
}

// Create a new Firefly with the given configuration.
func NewFirefly(baseUrl string, opts ...FireflyOption) *Firefly {
	var options fireflyOpts
	for _, opt := range opts {
		err := opt(&options)
		assert.NoError(err, "Error applying Firefly option")
	}

	if options.timeout == 0 {
		options.timeout = defaultTimeout
	}

	return &Firefly{
		baseUrl: baseUrl,
		httpClient: &http.Client{
			Timeout: options.timeout,
		},
		fireflyOpts: options,
	}
}

const defaultTimeout = 10 * time.Second

type fireflyOpts struct {
	apiKey  *string
	timeout time.Duration
}

// FireflyOption is a function that updates the fireflyOpts struct.
type FireflyOption func(*fireflyOpts) error

// WithApiKey is a configuration function that updates the api key used for each request.
func WithApiKey(apiKey string) FireflyOption {
	return func(c *fireflyOpts) error {
		trim := strings.TrimSpace(apiKey)
		if trim == "" {
			return ErrFireflyEmptyApiKey
		}
		c.apiKey = &trim
		return nil
	}
}

// WithTimeout is a configuration function that updates the client timeout.
func WithTimeout(timeout time.Duration) FireflyOption {
	return func(c *fireflyOpts) error {
		c.timeout = timeout
		return nil
	}
}

// addHeaders adds the required headers to the request.
func (f *Firefly) addHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *f.apiKey))
	req.Header.Set("Content-Type", "application/json")
}

// handleHttpErrorResponse will read the response body and return a FireflyErrReply.
func (f *Firefly) handleHttpErrorResponse(r *http.Response) error {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	var res models.FireflyErrReply
	err = json.Unmarshal(data, &res)
	if err != nil {
		return err
	}

	res.Code = r.StatusCode
	res.Status = r.Status

	return res
}

// CreateTransaction will create a new transaction in Firefly III.
func (f *Firefly) CreateTransaction(t *models.StoreTransactionRequest) error {
	url := fmt.Sprintf("%s/api/v1/transactions", f.baseUrl)
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	f.addHeaders(req)
	r, err := f.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return f.handleHttpErrorResponse(r)
	}

	_, err = io.ReadAll(r.Body)
	return err
}

// UpdateTransaction will create a new transaction in Firefly III.
func (f *Firefly) UpdateTransaction(id int, t *models.UpdateTransactionRequest) error {
	url := fmt.Sprintf("%s/api/v1/transactions/%d", f.baseUrl, id)
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	f.addHeaders(req)
	r, err := f.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return f.handleHttpErrorResponse(r)
	}

	_, err = io.ReadAll(r.Body)
	return err
}
