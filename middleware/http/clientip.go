package http

import (
	"context"
	"net"
	"net/http"
	"strings"
)

type ContextKey uint

const (
	ClientIP ContextKey = iota
)

type Provider uint8

const (
	NotProvided Provider = iota
	Cloudflare
	CloudFront
)

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
