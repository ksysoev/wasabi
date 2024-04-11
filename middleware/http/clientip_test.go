package http

import (
	"net/http"
	"testing"
)

func TestGetIPFromRequest(t *testing.T) {
	// Test with Cloudflare provider
	r, _ := http.NewRequest("GET", "http://example.com", http.NoBody)
	r.Header.Set("CF-Connecting-IP", "192.168.0.2")

	ip := getIPFromRequest(Cloudflare, r)

	if ip != "192.168.0.2" {
		t.Errorf("Expected IP to be 192.168.0.2, but got %s", ip)
	}

	r.Header.Set("True-Client-IP", "192.168.0.1")
	ip = getIPFromRequest(Cloudflare, r)

	if ip != "192.168.0.1" {
		t.Errorf("Expected IP to be 192.168.0.1, but got %s", ip)
	}

	// Test with CloudFront provider
	r, _ = http.NewRequest("GET", "http://example.com", http.NoBody)
	r.Header.Set("CloudFront-Viewer-Address", "192.168.0.3:1234")
	ip = getIPFromRequest(CloudFront, r)

	if ip != "192.168.0.3" {
		t.Errorf("Expected IP to be 192.168.0.3, but got %s", ip)
	}

	// Test with X-Real-Ip header
	r, _ = http.NewRequest("GET", "http://example.com", http.NoBody)
	r.Header.Set("X-Real-Ip", "192.168.0.4")

	ip = getIPFromRequest(NotProvided, r)
	if ip != "192.168.0.4" {
		t.Errorf("Expected IP to be 192.168.0.4, but got %s", ip)
	}

	// Test with X-Forwarded-For header
	r, _ = http.NewRequest("GET", "http://example.com", http.NoBody)
	r.Header.Set("X-Forwarded-For", "192.168.0.5, 192.168.0.6")

	ip = getIPFromRequest(NotProvided, r)
	if ip != "192.168.0.5" {
		t.Errorf("Expected IP to be 192.168.0.5, but got %s", ip)
	}

	// Test with RemoteAddr
	r = &http.Request{
		RemoteAddr: "192.168.0.7:1234",
	}

	ip = getIPFromRequest(NotProvided, r)
	if ip != "192.168.0.7" {
		t.Errorf("Expected IP to be 192.168.0.7, but got %s", ip)
	}

	// Test with empty request
	r = &http.Request{}

	ip = getIPFromRequest(NotProvided, r)
	if ip != "" {
		t.Errorf("Expected IP to be empty, but got %s", ip)
	}
}
