package request

import (
	"testing"
	"time"
)

func TestGetRetryInterval_WithLinearRetryPolicy(t *testing.T) {
	interval := time.Microsecond
	expectedBackoff := []time.Duration{interval, interval, interval}

	for i, v := range expectedBackoff {
		actualBackoff := LinearGetRetryInterval(interval)(i)
		if actualBackoff != v {
			t.Errorf("Expected actual backoff to be %v, but got %v", v, actualBackoff)
		}
	}
}

func TestGetRetryInterval_WithExponentialRetryPolicy(t *testing.T) {
	interval := time.Microsecond
	expectedBackoff := []time.Duration{interval, 2 * interval, 4 * interval}
	delayFactor := 2

	for i, v := range expectedBackoff {
		actualBackoff := ExponentialGetRetryInterval(interval, delayFactor)(i)
		if actualBackoff != v {
			t.Errorf("Expected actual backoff to be %v, but got %v", v, actualBackoff)
		}
	}
}
