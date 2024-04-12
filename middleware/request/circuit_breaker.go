package request

import (
	"fmt"
	"sync"
	"time"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
)

type CircuitBreakerState uint8

var (
	ErrCircuitBreakerOpen = fmt.Errorf("circuit breaker is open")
)

const (
	Closed CircuitBreakerState = iota
	Open
)

// NewCircuitBreakerMiddleware creates a new circuit breaker middleware with the specified threshold and period.
// The circuit breaker middleware wraps a given `wasabi.RequestHandler` and provides circuit breaking functionality.
// The circuit breaker tracks the number of consecutive errors and opens the circuit when the error count exceeds the threshold.
// During the open state, all requests are rejected with an `ErrCircuitBreakerOpen` error.
// After a specified period of time, the circuit breaker transitions to the semi-open state, allowing a single request to be processed.
// If the request succeeds, the circuit breaker resets the error count and transitions back to the closed state.
// If the request fails, the circuit breaker remains in the open state.
// The circuit breaker uses synchronization primitives to ensure thread safety.
// The `treshold` parameter specifies the maximum number of consecutive errors allowed before opening the circuit.
// The `period` parameter specifies the duration of time after which the circuit breaker transitions to the semi-open state.
// The returned function is a middleware that can be used with the `wasabi` framework.
func NewCircuitBreakerMiddleware(treshold int, period time.Duration) func(next wasabi.RequestHandler) wasabi.RequestHandler {
	intervalEnds := time.Now().Add(period)
	state := Closed
	errorCounter := 0
	lock := &sync.RWMutex{}
	semiOpenLock := &sync.Mutex{}

	return func(next wasabi.RequestHandler) wasabi.RequestHandler {
		return dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
			lock.RLock()

			switch state {
			case Closed:
				err := next.Handle(conn, req)
				if err == nil {
					return nil
				}

				lock.RUnlock()
				lock.Lock()
				now := time.Now()
				if intervalEnds.Before(time.Now()) {
					intervalEnds = now.Add(period)
					errorCounter = 0
				}

				errorCounter++
				if errorCounter >= treshold {
					state = Open
				}
				lock.Unlock()
				return err
			case Open:
				lock.RUnlock()
				if !semiOpenLock.TryLock() {
					return ErrCircuitBreakerOpen
				}

				defer semiOpenLock.Unlock()

				err := next.Handle(conn, req)
				if err == nil {
					lock.Lock()
					state = Closed
					errorCounter = 0
				}

				return err
			default:
				panic("Unknown state of circuit breaker")
			}
		})
	}
}
