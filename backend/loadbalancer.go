package backend

import (
	"fmt"
	"math"
	"sync/atomic"

	"github.com/ksysoev/wasabi"
)

var ErrNotEnoughBackends = fmt.Errorf("load balancer requires at least 2 backends")

const (
	minRequiredBackends = 2
	errorThreshold      = 5
)

type LoadBalancerNode struct {
	backend       wasabi.RequestHandler
	counter       atomic.Int32
	errors        atomic.Int32
	alive         atomic.Bool
	lastLiveCheck atomic.Value
}

type LoadBalancer struct {
	backends []*LoadBalancerNode
}

// NewLoadBalancer creates a new instance of LoadBalancer with the given backends.
// It takes a slice of RequestHandler as a parameter and returns a new instance of LoadBalancer.
func NewLoadBalancer(backends []wasabi.RequestHandler) (*LoadBalancer, error) {
	if len(backends) < minRequiredBackends {
		return nil, ErrNotEnoughBackends
	}

	nodes := make([]*LoadBalancerNode, len(backends))

	for i, backend := range backends {
		nodes[i] = &LoadBalancerNode{
			backend:       backend,
			counter:       atomic.Int32{},
			errors:        atomic.Int32{},
			alive:         atomic.Bool{},
			lastLiveCheck: atomic.Value{},
		}

		nodes[i].alive.Store(true)
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

	err := backend.backend.Handle(conn, r)
	if err != nil {
		backend.errors.Add(1)
		if backend.errors.Load() > errorThreshold {
			backend.alive.Store(false)
		}
		return err
	}

	if backend.errors.Load() > 0 {
		backend.errors.Store(0)
		backend.alive.Store(true)
	}

	return nil
}

// getLeastBusyNode returns the least busy backend node.
// It returns the least busy backend node.
func (lb *LoadBalancer) getLeastBusyNode() *LoadBalancerNode {
	var minBackend *LoadBalancerNode
	var minRequests int32 = math.MaxInt32

	allDown := true
	for _, b := range lb.backends {
		if b.alive.Load() {
			allDown = false
			counter := b.counter.Load()

			if counter < minRequests {
				minRequests = counter
				minBackend = b
			}
		}
	}

	if allDown {
		for _, b := range lb.backends {
			counter := b.counter.Load()
			if minBackend == nil || counter < minRequests {
				minRequests = counter
				minBackend = b
			}
		}
	}

	return minBackend
}
