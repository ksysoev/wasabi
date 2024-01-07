package wasabi

import (
	"bytes"
	"log/slog"
	"net/http"
)

type Backend interface {
	Handle(conn *Connection, r Request) error
}

type HttpBackend struct {
	endpoint string
}

func NewBackend(endpoint string) *HttpBackend {
	return &HttpBackend{endpoint: endpoint}
}

func (b *HttpBackend) Handle(conn *Connection, r Request) error {
	req, ok := r.(*JSONRPCRequest)
	if !ok {
		return nil
	}
	body := bytes.NewBufferString(req.Data)
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
	apiResp := NewResponse(req.ID, respBody.String())
	data, err := apiResp.String()

	if err != nil {
		slog.Error("Error creating response", "error", err)
		return err
	}

	return conn.SendResponse(data)
}
