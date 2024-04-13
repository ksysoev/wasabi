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

// NewCircuitBreakerMiddleware creates a new circuit breaker middleware with the specified parameters.
// It returns a function that wraps the provided `wasabi.RequestHandler` and implements the circuit breaker logic.
// The circuit breaker monitors the number of errors and successful requests within a given time period.
// If the number of errors exceeds the threshold, the circuit breaker switches to the "Open" state and rejects subsequent requests.
// After a specified number of successful requests, the circuit breaker switches back to the "Closed" state.
// The circuit breaker uses a lock to ensure thread safety.
// The `treshold` parameter specifies the maximum number of errors allowed within the time period.
// The `period` parameter specifies the duration of the time period.
// The `recoverAfter` parameter specifies the number of successful requests required to recover from the "Open" state.
// The returned function can be used as middleware in a Wasabi server.
func NewCircuitBreakerMiddleware(treshold uint, period time.Duration, recoverAfter uint) func(next wasabi.RequestHandler) wasabi.RequestHandler {
	var errorCounter, successCounter uint

	intervalEnds := time.Now().Add(period)
	state := Closed

	lock := &sync.RWMutex{}
	semiOpenLock := &sync.Mutex{}

	return func(next wasabi.RequestHandler) wasabi.RequestHandler {
		return dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
			lock.RLock()
			currentState := state
			lock.RUnlock()

			switch currentState {
			case Closed:
				err := next.Handle(conn, req)
				if err == nil {
					return nil
				}

				lock.Lock()
				defer lock.Unlock()

				now := time.Now()
				if intervalEnds.Before(time.Now()) {
					intervalEnds = now.Add(period)
					errorCounter = 0
				}

				errorCounter++
				if errorCounter >= treshold {
					state = Open
				}

				return err
			case Open:
				if !semiOpenLock.TryLock() {
					return ErrCircuitBreakerOpen
				}

				defer semiOpenLock.Unlock()

				err := next.Handle(conn, req)

				lock.Lock()
				defer lock.Unlock()

				if err != nil {
					successCounter = 0
					return err
				}

				successCounter++

				if successCounter >= recoverAfter {
					state = Closed
					errorCounter = 0
					successCounter = 0
				}

				return nil
			default:
				panic("Unknown state of circuit breaker")
			}
		})
	}
}
