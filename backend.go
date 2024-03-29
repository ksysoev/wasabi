package wasabi

import (
	"bytes"
	"log/slog"
	"net/http"
)

type Backend interface {
	Handle(conn Connection, r Request) error
}

type HTTPBackend struct {
	endpoint string
}

func NewBackend(endpoint string) *HTTPBackend {
	return &HTTPBackend{endpoint: endpoint}
}

func (b *HTTPBackend) Handle(conn Connection, r Request) error {
	req := string(r.Data())
	body := bytes.NewBufferString(req)
	httpReq, err := http.NewRequest("POST", b.endpoint, body)

	if err != nil {
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(httpReq)

	if err != nil {
		slog.Error("Error sending request", "error", err)
		return err
	}
	defer resp.Body.Close()

	respBody := bytes.NewBuffer(make([]byte, 0))

	_, err = respBody.ReadFrom(resp.Body)
	if err != nil {
		slog.Error("Error reading response body", "error", err)
		return err
	}

	return conn.Send(respBody.Bytes())
}
