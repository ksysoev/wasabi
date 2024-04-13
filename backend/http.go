package backend

import (
	"bytes"
	"net/http"
	"time"

	"github.com/ksysoev/wasabi"
)

const defaultTimeout = 30 * time.Second

// HTTPBackend represents an HTTP backend for handling requests.
type HTTPBackend struct {
	factory RequestFactory
	client  *http.Client
}

type httpBackendConfig struct {
	defaultTimeout time.Duration
}

type HTTPBackendOption func(*httpBackendConfig)

// NewBackend creates a new instance of HTTPBackend with the given RequestFactory.
func NewBackend(factory RequestFactory, options ...HTTPBackendOption) *HTTPBackend {
	httpBackendConfig := &httpBackendConfig{
		defaultTimeout: defaultTimeout,
	}

	for _, option := range options {
		option(httpBackendConfig)
	}

	return &HTTPBackend{
		factory: factory,
		client: &http.Client{
			Timeout: httpBackendConfig.defaultTimeout,
		},
	}
}

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

	respBody := bytes.NewBuffer(make([]byte, 0))

	_, err = respBody.ReadFrom(resp.Body)
	if err != nil {
		return err
	}

	return conn.Send(wasabi.MsgTypeText, respBody.Bytes())
}

func WithDefaultHTTPTimeout(timeout time.Duration) HTTPBackendOption {
	return func(cfg *httpBackendConfig) {
		cfg.defaultTimeout = timeout
	}
}
