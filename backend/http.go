package backend

import (
	"bytes"
	"net/http"

	"github.com/ksysoev/wasabi"
)

// HTTPBackend represents an HTTP backend for handling requests.
type HTTPBackend struct {
	factory RequestFactory
	client  *http.Client
}

// NewBackend creates a new instance of HTTPBackend with the given RequestFactory.
func NewBackend(factory RequestFactory) *HTTPBackend {
	return &HTTPBackend{
		factory: factory,
		client:  &http.Client{},
	}
}

// Handle handles the incoming connection and request.
// It sends the request to the backend server and returns the response to the connection.
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
