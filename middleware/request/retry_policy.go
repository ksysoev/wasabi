package request

import (
	"math"
	"time"
)

// GetRetryInterval is a higher order function that returns a function to get retry interval.
// The inner function takes the retry interation (int) as an argument and returns the duration until next retry
type GetRetryInterval func(int) time.Duration

// ExponentialGetRetryInterval provides an exponential backoff implementation of GetRetryInterval.
// It enables the retry middleware to ensure exponentially higher wait times between retries.
// Parameters:
// - seed: The initial duration to wait until exponential backoff is triggered
// - delayFactor: The exponential index for backoff. Higher the delay factor, larger is the time between retries
//
// Returns:
// - GetRetryInterval: returns the backoff function to get interval delay for current iteration
func ExponentialGetRetryInterval(seed time.Duration, delayFactor int) GetRetryInterval {
	return func(iteration int) time.Duration {
		return seed * (time.Duration(math.Pow(float64(delayFactor), float64(iteration))))
	}
}

// LinearGetRetryInterval is an linear delay implementation of GetRetryInterval.
// It enables the retry middleware to ensure constant wait times between retries.
// Parameters:
// - interval: The duration to wait between retries
//
// Returns:
// - GetRetryInterval: returns the backoff function to get interval delay for current iteration
func LinearGetRetryInterval(interval time.Duration) GetRetryInterval {
	return func(_ int) time.Duration {
		return interval
	}
}
