package request

import (
	"testing"
	"time"
)

func TestGetRetryInterval_WithLinearRetryPolicy(t *testing.T) {
	interval := time.Microsecond
	expectedBackoff := []time.Duration{interval, interval, interval}

	for i, v := range expectedBackoff {
		actualBackoff := LinearRetryConfig(len(expectedBackoff), interval).getRetryInterval(i)
		if actualBackoff != v {
			t.Errorf("Expected actual backoff to be %v, but got %v", v, actualBackoff)
		}
	}
}

func TestGetRetryInterval_WithExponentialRetryPolicy(t *testing.T) {
	interval := time.Microsecond
	expectedBackoff := []time.Duration{interval, 2 * interval, 4 * interval}

	for i, v := range expectedBackoff {
		actualBackoff := ExponentialRetryConfig(len(expectedBackoff), interval, 2).getRetryInterval(i)
		if actualBackoff != v {
			t.Errorf("Expected actual backoff to be %v, but got %v", v, actualBackoff)
		}
	}
}
