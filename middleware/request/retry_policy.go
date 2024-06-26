package request

import (
	"fmt"
	"math"
	"time"
)

type RetryPolicy int

const (
	LinearRetryPolicy RetryPolicy = iota + 1
	ExponentialRetryPolicy
)

const defaultDelayFactor = 2

type RetryConfig struct {
	retryPolicy  RetryPolicy
	maxRetries   int
	seedInterval time.Duration
	delayFactor  int
}

func LinearRetryConfig(maxRetries int, interval time.Duration) *RetryConfig {
	return &RetryConfig{LinearRetryPolicy, maxRetries, interval, -1}
}

func ExponentialRetryConfig(maxRetries int, seed time.Duration, delayFactor int) *RetryConfig {
	return &RetryConfig{ExponentialRetryPolicy, maxRetries, seed, delayFactor}
}

func ExponentialRetryConfigWithDefaultDelayFactor(maxRetries int, seed time.Duration) *RetryConfig {
	return &RetryConfig{ExponentialRetryPolicy, maxRetries, seed, defaultDelayFactor}
}

func GetRetryInterval(policy RetryPolicy, seed time.Duration, iteration, delayFactor int) time.Duration {
	switch policy {
	case LinearRetryPolicy:
		return seed
	case ExponentialRetryPolicy:
		return seed * (time.Duration(math.Pow(float64(delayFactor), float64(iteration))))
	default:
		errMsg := fmt.Sprintf("Unsupported retry policy %v", policy)
		panic(errMsg)
	}
}
