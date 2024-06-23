package backend

import (
	"testing"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/mocks"
)

func TestNewLoadBalancer(t *testing.T) {
	backends := []wasabi.RequestHandler{
		mocks.NewMockBackend(t),
	}

	_, err := NewLoadBalancer(backends)
	if err != ErrNotEnoughBackends {
		t.Errorf("Expected error to be 'load balancer requires at least 2 backends', but got %v", err)
	}

	backends = append(backends, mocks.NewMockBackend(t))

	lb, err := NewLoadBalancer(backends)
	if err != nil {
		t.Fatalf("Failed to create load balancer: %v", err)
	}

	if len(lb.backends) != len(backends) {
		t.Errorf("Expected %d backends, but got %d", len(backends), len(lb.backends))
	}

	for i, backend := range lb.backends {
		if backend.backend != backends[i] {
			t.Errorf("Expected backend at index %d to be %v, but got %v", i, backends[i], backend.backend)
		}

		if backend.counter.Load() != 0 {
			t.Errorf("Expected backend counter at index %d to be 0, but got %d", i, backend.counter.Load())
		}
	}
}

func TestLoadBalancer_getLeastBusyNode(t *testing.T) {
	backends := []wasabi.RequestHandler{
		// Mock backends for testing
		mocks.NewMockBackend(t),
		mocks.NewMockBackend(t),
	}

	lb, err := NewLoadBalancer(backends)
	if err != nil {
		t.Fatalf("Failed to create load balancer: %v", err)
	}

	// Increment the counter of the second backend
	lb.backends[0].counter.Add(10)

	// Get the least busy node
	leastBusyNode := lb.getLeastBusyNode()

	// Check if the least busy node is the second backend
	if leastBusyNode != lb.backends[1] {
		t.Errorf("Expected least busy node to be the second backend, but got %v", leastBusyNode)
	}
}
func TestLoadBalancer_Handle(t *testing.T) {
	firstBackend := mocks.NewMockBackend(t)
	// Create mock backends for testing
	backends := []wasabi.RequestHandler{
		firstBackend,
		mocks.NewMockBackend(t),
		mocks.NewMockBackend(t),
	}

	lb, err := NewLoadBalancer(backends)
	if err != nil {
		t.Fatalf("Failed to create load balancer: %v", err)
	}

	// Create mock connection and request
	mockConn := mocks.NewMockConnection(t)
	mockRequest := mocks.NewMockRequest(t)

	firstBackend.EXPECT().Handle(mockConn, mockRequest).Return(nil)

	// Call the Handle method
	err = lb.Handle(mockConn, mockRequest)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
