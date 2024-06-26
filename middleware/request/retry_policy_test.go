package request

import (
	"testing"
	"time"
)

func TestGetRetryInterval_WithLinearRetryPolicy(t *testing.T) {
	interval := time.Microsecond
	expectedBackoff := []time.Duration{time.Microsecond, time.Microsecond, time.Microsecond}

	for i, v := range expectedBackoff {
		actualBackoff := GetRetryInterval(LinearRetryPolicy, interval, i, -1)
		if actualBackoff != v {
			t.Errorf("Expected actual backoff to be %v, but got %v", v, actualBackoff)
		}
	}
}

func TestGetRetryInterval_WithExponentialRetryPolicy(t *testing.T) {
	interval := time.Microsecond
	expectedBackoff := []time.Duration{time.Microsecond, 2 * time.Microsecond, 4 * time.Microsecond}

	for i, v := range expectedBackoff {
		actualBackoff := GetRetryInterval(ExponentialRetryPolicy, interval, i, 2)
		if actualBackoff != v {
			t.Errorf("Expected actual backoff to be %v, but got %v", v, actualBackoff)
		}
	}
}
