package backend

import (
	"bytes"
	"net/http"

	"github.com/ksysoev/wasabi"
)

type HTTPBackend struct {
	factory RequestFactory
	client  *http.Client
}

func NewBackend(factory RequestFactory) *HTTPBackend {
	return &HTTPBackend{
		factory: factory,
		client:  &http.Client{},
	}
}

func (b *HTTPBackend) Handle(conn wasabi.Connection, r wasabi.Request) error {
	httpReq, err := b.factory(r)
	if err != nil {
		return err
	}

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

	return conn.Send(respBody.String())
}
