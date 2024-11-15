package backend

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/channel"
)

const defaultTimeout = 30 * time.Second
const defaultMaxReqPerHost = 50

// HTTPBackend represents an HTTP backend for handling requests.
type HTTPBackend struct {
	factory RequestFactory
	client  *http.Client
}

type httpBackendConfig struct {
	defaultTimeout time.Duration
	maxReqPerHost  int
}

type HTTPBackendOption func(*httpBackendConfig)

// NewBackend creates a new instance of HTTPBackend with the given RequestFactory.
func NewBackend(factory RequestFactory, options ...HTTPBackendOption) *HTTPBackend {
	httpBackendConfig := &httpBackendConfig{
		defaultTimeout: defaultTimeout,
		maxReqPerHost:  defaultMaxReqPerHost,
	}

	for _, option := range options {
		option(httpBackendConfig)
	}

	return &HTTPBackend{
		factory: factory,
		client: &http.Client{
			Timeout: httpBackendConfig.defaultTimeout,
			Transport: &http.Transport{
				MaxConnsPerHost: httpBackendConfig.maxReqPerHost,
			},
		},
	}
}

// Handle handles the incoming request by sending it to the HTTP server and returning the response.
// It takes a connection and a request as parameters and returns an error if any.
func (b *HTTPBackend) Handle(conn wasabi.Connection, r wasabi.Request) error {
	httpReq, err := b.factory(r)
	if err != nil {
		return err
	}

	httpReq = httpReq.WithContext(r.Context())

	resp, err := b.client.Do(httpReq)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := conn.Send(wasabi.MsgTypeText, body); err != nil {
		if err == channel.ErrConnectionClosed {
			return nil
		}

		return err
	}

	return nil
}

// WithTimeout sets the default timeout for the HTTP client.
func WithTimeout(timeout time.Duration) HTTPBackendOption {
	return func(cfg *httpBackendConfig) {
		cfg.defaultTimeout = timeout
	}
}

// WithMaxRequestsPerHost sets the maximum number of requests per host.
func WithMaxRequestsPerHost(maxReqPerHost int) HTTPBackendOption {
	return func(cfg *httpBackendConfig) {
		cfg.maxReqPerHost = maxReqPerHost
	}
}
