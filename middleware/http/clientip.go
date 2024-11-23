package http

import (
	"context"
	"net"
	"net/http"
	"strings"
)

type Provider uint8

const (
	NotProvided Provider = iota
	Cloudflare
	CloudFront
)

// NewClientIPMiddleware returns a middleware function that extracts the client's IP address from the request
// and adds it to the request's context. The IP address is obtained using the provided IP address provider.
// The middleware function takes the next http.Handler as input and returns a new http.Handler that wraps
// the provided handler. When the new handler is called, it first extracts the client's IP address from the
// request using the provider, adds it to the request's context, and then calls the next handler in the chain.
// The IP address can be accessed from the request's context using the ClientIP key.
func NewClientIPMiddleware(provider Provider) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := getIPFromRequest(provider, r)

			ctx := r.Context()
			ctx = context.WithValue(ctx, ClientIP, ip)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// GetClientIP retrieves the client IP address from the provided context.
// It takes a single parameter ctx of type context.Context.
// It returns a string representing the client IP address if found, otherwise an empty string.
func GetClientIP(ctx context.Context) string {
	if ip, ok := ctx.Value(ClientIP).(string); ok {
		return ip
	}

	return ""
}

func getIPFromRequest(provider Provider, r *http.Request) string {
	var ip string

	switch provider {
	case Cloudflare:
		ip = r.Header.Get("True-Client-IP")
		if ip == "" {
			ip = r.Header.Get("CF-Connecting-IP")
		}
	case CloudFront:
		ip = r.Header.Get("CloudFront-Viewer-Address")
		if ip != "" {
			parts := strings.Split(ip, ":")
			if len(parts) > 0 {
				ip = parts[0]
			}
		}
	default:
	}

	if ip != "" {
		return ip
	}

	if ip := r.Header.Get("X-Real-Ip"); ip != "" {
		return ip
	}

	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		parts := strings.Split(ip, ",")
		if len(parts) > 0 {
			return parts[0]
		}
	}

	if ip, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return ip
	}

	return ""
}
