package backend

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ksysoev/wasabi"
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

	length := 0
	if resp.ContentLength > 0 {
		length = int(resp.ContentLength)
	} else {
		return fmt.Errorf("response content length is unknown")
	}

	body := make([]byte, length)
	_, err = resp.Body.Read(body)

	if err != nil && err != io.EOF {
		return err
	}

	return conn.Send(wasabi.MsgTypeText, body)
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
