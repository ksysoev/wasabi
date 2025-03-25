package backend

import (
	"testing"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/mocks"
)

func TestNewLoadBalancer(t *testing.T) {
	backends := []struct {
		Handler wasabi.RequestHandler
		Weight  int32
	}{{
		Handler: mocks.NewMockBackend(t),
		Weight:  1,
	}}

	_, err := NewLoadBalancer(backends)
	if err != ErrNotEnoughBackends {
		t.Errorf("Expected error to be 'load balancer requires at least 2 backends', but got %v", err)
	}

	backends = append(backends, struct {
		Handler wasabi.RequestHandler
		Weight  int32
	}{mocks.NewMockBackend(t), 1})

	lb, err := NewLoadBalancer(backends)
	if err != nil {
		t.Fatalf("Failed to create load balancer: %v", err)
	}

	if len(lb.backends) != len(backends) {
		t.Errorf("Expected %d backends, but got %d", len(backends), len(lb.backends))
	}

	for i, backend := range lb.backends {
		if backend.backend != backends[i].Handler {
			t.Errorf("Expected backend at index %d to be %v, but got %v", i, backends[i], backend.backend)
		}

		if backend.counter.Load() != 0 {
			t.Errorf("Expected backend counter at index %d to be 0, but got %d", i, backend.counter.Load())
		}
	}
}

func TestLoadBalancer_getLeastBusyNode(t *testing.T) {
	backends := []struct {
		Handler wasabi.RequestHandler
		Weight  int32
	}{{
		Handler: mocks.NewMockBackend(t),
		Weight:  1,
	}, {
		Handler: mocks.NewMockBackend(t),
		Weight:  1,
	}}

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

func TestLoadBalancer_getLeastBusyNodeWeighted(t *testing.T) {
	backends := []struct {
		Handler wasabi.RequestHandler
		Weight  int32
	}{{
		Handler: mocks.NewMockBackend(t),
		Weight:  1,
	}, {
		Handler: mocks.NewMockBackend(t),
		Weight:  3, // supposed to handle 3x load to be as busy as the second handler
	}}

	lb, err := NewLoadBalancer(backends)
	if err != nil {
		t.Fatalf("Failed to create load balancer: %v", err)
	}

	// Increment the counter of the backends
	lb.backends[0].counter.Add(1)
	lb.backends[1].counter.Add(2)

	// Get the least busy node
	leastBusyNode := lb.getLeastBusyNode()

	// Check if the least busy node is the second backend
	if leastBusyNode != lb.backends[1] {
		t.Errorf("Expected least busy node to be the second backend, but got %v because the counter of first backend is: %d and counter of second backend is: %d", leastBusyNode, lb.backends[0].counter.Load(), lb.backends[1].counter.Load())
	}
}

func TestLoadBalancer_getLeastBusyNodeZeroWeight(t *testing.T) {
	backends := []struct {
		Handler wasabi.RequestHandler
		Weight  int32
	}{{
		Handler: mocks.NewMockBackend(t),
		Weight:  0,
	}, {
		Handler: mocks.NewMockBackend(t),
		Weight:  1,
	}}

	lb, err := NewLoadBalancer(backends)
	if err != nil {
		t.Fatalf("Failed to create load balancer: %v", err)
	}

	// Get the least busy node
	leastBusyNode := lb.getLeastBusyNode()

	// Check if the least busy node is the second backend
	if leastBusyNode != lb.backends[1] {
		t.Errorf("Expected least busy node to be the second backend, but got %v", leastBusyNode)
	}
}

func TestLoadBalancer_Handle(t *testing.T) {
	mockBackend := mocks.NewMockBackend(t)

	firstBackend := struct {
		Handler wasabi.RequestHandler
		Weight  int32
	}{Handler: mockBackend, Weight: 1}

	// Create mock backends for testing
	backends := []struct {
		Handler wasabi.RequestHandler
		Weight  int32
	}{firstBackend, {
		Handler: mocks.NewMockBackend(t),
		Weight:  1,
	}, {
		Handler: mocks.NewMockBackend(t),
		Weight:  1,
	}}

	lb, err := NewLoadBalancer(backends)
	if err != nil {
		t.Fatalf("Failed to create load balancer: %v", err)
	}

	// Create mock connection and request
	mockConn := mocks.NewMockConnection(t)
	mockRequest := mocks.NewMockRequest(t)

	mockBackend.EXPECT().Handle(mockConn, mockRequest).Return(nil)

	// Call the Handle method
	err = lb.Handle(mockConn, mockRequest)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
