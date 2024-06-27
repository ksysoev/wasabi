package request

import (
	"math"
	"time"
)

type GetRetryInterval func(int) time.Duration

func ExponentialGetRetryInterval(seed time.Duration, delayFactor int) GetRetryInterval {
	return func(iteration int) time.Duration {
		return seed * (time.Duration(math.Pow(float64(delayFactor), float64(iteration))))
	}
}

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
