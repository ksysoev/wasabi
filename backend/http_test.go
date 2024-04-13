package backend

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/mocks"
)

func TestNewBackend(t *testing.T) {
	factory := func(req wasabi.Request) (*http.Request, error) {
		return nil, nil
	}

	backend := NewBackend(factory)
	if backend.client.Timeout != defaultTimeout {
		t.Errorf("Expected default timeout to be %v, but got %v", defaultTimeout, backend.client.Timeout)
	}

	backend = NewBackend(factory, WithDefaultHTTPTimeout(10))

	if backend.client.Timeout != 10 {
		t.Errorf("Expected default timeout to be 10, but got %v", backend.client.Timeout)
	}
}

func TestHTTPBackend_Handle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`OK`))
	}))
	defer server.Close()

	mockConn := mocks.NewMockConnection(t)
	mockReq := mocks.NewMockRequest(t)

	mockReq.EXPECT().Context().Return(context.Background())

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

func TestHTTPBackend_Handle_ErrorCreatingHTTPRequest(t *testing.T) {
	testError := errors.New("test error")

	backend := NewBackend(func(req wasabi.Request) (*http.Request, error) {
		return nil, testError
	})

	mockConn := mocks.NewMockConnection(t)
	mockReq := mocks.NewMockRequest(t)

	err := backend.Handle(mockConn, mockReq)

	if err != testError {
		t.Errorf("Expected error to be %v, but got %v", testError, err)
	}
}

func TestHTTPBackend_Handle_TimeoutRequestByContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep for 1 second
		<-time.After(10 * time.Millisecond)
		_, _ = w.Write([]byte(`OK`))
	}))
	defer server.Close()

	mockConn := mocks.NewMockConnection(t)
	mockReq := mocks.NewMockRequest(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	mockReq.EXPECT().Context().Return(ctx)

	backend := NewBackend(func(req wasabi.Request) (*http.Request, error) {
		bodyReader := bytes.NewBufferString("test request")
		httpReq, err := http.NewRequest("GET", server.URL, bodyReader)
		if err != nil {
			return nil, err
		}
		return httpReq, nil
	})

	err := backend.Handle(mockConn, mockReq)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected error to be %v, but got %v", context.DeadlineExceeded, err)
	}
}
