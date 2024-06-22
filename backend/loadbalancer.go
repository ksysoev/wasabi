package backend

import (
	"sync/atomic"

	"github.com/ksysoev/wasabi"
)

type BackendNode struct {
	backend wasabi.RequestHandler
	counter atomic.Int32
}

type LoadBalancer struct {
	backends []*BackendNode
}

// NewLoadBalancer creates a new instance of LoadBalancer with the given backends.
// It takes a slice of RequestHandler as a parameter and returns a new instance of LoadBalancer.
func NewLoadBalancer(backends []wasabi.RequestHandler) *LoadBalancer {
	nodes := make([]*BackendNode, len(backends))

	for i, backend := range backends {
		nodes[i] = &BackendNode{
			backend: backend,
			counter: atomic.Int32{},
		}
	}

	return &LoadBalancer{
		backends: nodes,
	}
}

// Handle handles the incoming request by sending it to the least busy backend and returning the response.
// It takes a connection and a request as parameters and returns an error if any.
func (lb *LoadBalancer) Handle(conn wasabi.Connection, r wasabi.Request) error {
	backend := lb.getLeastBusyNode()

	backend.counter.Add(1)
	defer backend.counter.Add(-1)

	return backend.backend.Handle(conn, r)
}

// getLeastBusyNode returns the least busy backend node.
// It returns the least busy backend node.
func (lb *LoadBalancer) getLeastBusyNode() *BackendNode {
	var minRequests int32
	var minBackend *BackendNode

	for i := range lb.backends {
		requests := lb.backends[i].counter.Load()

		if requests < minRequests {
			minRequests = requests
			minBackend = lb.backends[i]
		}
	}

	return minBackend
}
