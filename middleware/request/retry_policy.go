package request

import (
	"math"
	"time"
)

// GetRetryInterval is a higher order function that returns a function to get retry interval.
// The inner function takes the retry interation (int) as an argument and returns the duration until next retry
type GetRetryInterval func(int) time.Duration

// ExponentialGetRetryInterval is an exponential backoff implementation of GetRetryInterval.
// It enables the retry middleware to ensure exponentially higher wait times between retries.
func ExponentialGetRetryInterval(seed time.Duration, delayFactor int) GetRetryInterval {
	return func(iteration int) time.Duration {
		return seed * (time.Duration(math.Pow(float64(delayFactor), float64(iteration))))
	}
}

// LinearGetRetryInterval is an linear delay implementation of GetRetryInterval.
// It enables the retry middleware to ensure constant wait times between retries.
func LinearGetRetryInterval(interval time.Duration) GetRetryInterval {
	return func(_ int) time.Duration {
		return interval
	}
}

const defaultDelayFactor = 2

type RetryConfig struct {
	getRetryInterval GetRetryInterval
	maxRetries       int
}

func LinearRetryConfig(maxRetries int, interval time.Duration) *RetryConfig {
	return &RetryConfig{LinearGetRetryInterval(interval), maxRetries}
}

func ExponentialRetryConfig(maxRetries int, seed time.Duration, delayFactor int) *RetryConfig {
	return &RetryConfig{ExponentialGetRetryInterval(seed, delayFactor), maxRetries}
}

func ExponentialRetryConfigWithDefaultDelayFactor(maxRetries int, seed time.Duration) *RetryConfig {
	return &RetryConfig{ExponentialGetRetryInterval(seed, defaultDelayFactor), maxRetries}
}
