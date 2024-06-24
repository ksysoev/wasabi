package request

import (
	"fmt"
	"time"

	"github.com/ksysoev/wasabi"
	"github.com/ksysoev/wasabi/dispatch"
	"github.com/sony/gobreaker/v2"
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
// After a set amount of period, the circuit breaker switches to the
// "Semi-open" state.
// If request succeeds in "Semi-open" state, the state will be changed to
// "Closed", else back to "Open".
// The `threshold` parameter specifies the maximum number of errors allowed within the time period.
// The `period` parameter specifies the duration of the time period.
// The returned function can be used as middleware in a Wasabi server.
func NewCircuitBreakerMiddleware(threshold uint, period time.Duration) func(next wasabi.RequestHandler) wasabi.RequestHandler {
	var st gobreaker.Settings
	st.Timeout = period
	st.ReadyToTrip = func(counts gobreaker.Counts) bool {
		return counts.ConsecutiveFailures >= uint32(threshold)
	}
	cb := gobreaker.NewCircuitBreaker[any](st)

	return func(next wasabi.RequestHandler) wasabi.RequestHandler {
		return dispatch.RequestHandlerFunc(func(conn wasabi.Connection, req wasabi.Request) error {
			_, err := cb.Execute(func() (any, error) {
				err := next.Handle(conn, req)
				if err != nil {
					return nil, err
				}
				return struct{}{}, nil
			})
			if err != nil {
				return err
			}
			return nil
		})
	}

}
