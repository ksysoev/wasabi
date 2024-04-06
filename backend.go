package wasabi

import (
	"bytes"
	"net/http"
)

type RequestFactory func(req Request) (*http.Request, error)

type HTTPBackend struct {
	factory RequestFactory
	client  *http.Client
	sem     chan struct{}
}

const (
	MaxConcurrentRequests = 50
)

func NewBackend(factory RequestFactory) *HTTPBackend {
	return &HTTPBackend{
		factory: factory,
		client:  &http.Client{},
		sem:     make(chan struct{}, MaxConcurrentRequests),
	}
}

func (b *HTTPBackend) Handle(conn Connection, r Request) error {
	httpReq, err := b.factory(r)
	if err != nil {
		return err
	}

	ctx := r.Context()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case b.sem <- struct{}{}:
	}

	resp, err := b.client.Do(httpReq)
	<-b.sem

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	respBody := bytes.NewBuffer(make([]byte, 0))

	_, err = respBody.ReadFrom(resp.Body)
	if err != nil {
		return err
	}

	return conn.Send(respBody.Bytes())
}
