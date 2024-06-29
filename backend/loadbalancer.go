package backend

import (
	"fmt"
	"math"
	"sync/atomic"

	"github.com/ksysoev/wasabi"
)

var ErrNotEnoughBackends = fmt.Errorf("load balancer requires at least 2 backends")

const minRequiredBackends = 2

type LoadBalancerNode struct {
	backend wasabi.RequestHandler
	counter atomic.Int32
	weight  uint
}

type LoadBalancer struct {
	backends []*LoadBalancerNode
}

// NewLoadBalancer creates a new instance of LoadBalancer with the given backends.
// It takes a slice of RequestHandler as a parameter and returns a new instance of LoadBalancer.
func NewLoadBalancer(backends []struct {
	handler wasabi.RequestHandler
	weight  uint
}) (*LoadBalancer, error) {
	if len(backends) < minRequiredBackends {
		return nil, ErrNotEnoughBackends
	}

	nodes := make([]*LoadBalancerNode, len(backends))

	for i, backend := range backends {
		nodes[i] = &LoadBalancerNode{
			backend: backend.handler,
			counter: atomic.Int32{},
			weight:  backend.weight,
		}
	}

	return &LoadBalancer{
		backends: nodes,
	}, nil
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
func (lb *LoadBalancer) getLeastBusyNode() *LoadBalancerNode {
	var minRequests = int32(math.MaxInt32)

	var minBackend *LoadBalancerNode

	for _, b := range lb.backends {
		if b.weight == 0 {
			continue
		}

		counter := b.counter.Load() / int32(b.weight)

		if counter < minRequests {
			minRequests = counter
			minBackend = b
		}
	}

	return minBackend
}
