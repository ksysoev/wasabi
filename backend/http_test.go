package backend

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/mocks"
)

func TestHTTPBackend_Handle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`OK`))
	}))
	defer server.Close()

	mockConn := mocks.NewMockConnection(t)
	mockReq := mocks.NewMockRequest(t)

	mockConn.EXPECT().Send("OK").Return(nil)
	mockReq.EXPECT().Data().Return([]byte("test request"))

	backend := NewBackend(func(req wasabi.Request) (*http.Request, error) {
		bodyReader := bytes.NewBufferString(string(req.Data()))
		httpReq, err := http.NewRequest("GET", server.URL, bodyReader)
		if err != nil {
			return nil, err
		}
		return httpReq, nil
	})

	err := backend.Handle(mockConn, mockReq)
	if err != nil {
		t.Fatal(err)
	}
}
