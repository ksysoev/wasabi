package server

import (
	"context"
	"time"
)

type WasabiContext interface {
	context.Context
	WithValue(key, value interface{}) WasabiContext
	GetHTTPConfig() HTTPConfig
}

type WasabiDefaultContextImpl struct {
	context.Context
}

func (w WasabiDefaultContextImpl) GetHTTPConfig() HTTPConfig {
	return *getDefaultHTTPConfig()
}

func (w WasabiDefaultContextImpl) WithValue(key, value interface{}) WasabiContext {
	return WasabiDefaultContextImpl{Context: context.WithValue(w.Context, key, value)}
}

type HTTPConfig struct {
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
}

func NewWasabiDefaultContext(fromContext context.Context) WasabiDefaultContextImpl {
	var ctx = WasabiDefaultContextImpl{fromContext}
	return ctx
}

var defaultHTTPConfig *HTTPConfig

func getDefaultHTTPConfig() *HTTPConfig {
	if defaultHTTPConfig == nil {
		defaultHTTPConfig = &HTTPConfig{
			ReadHeaderTimeout: ReadHeaderTimeoutSeconds * time.Second,
			ReadTimeout:       ReadTimeoutSeconds * time.Second,
		}
	}

	return defaultHTTPConfig
}

const (
	ReadHeaderTimeoutSeconds = 3
	ReadTimeoutSeconds       = 30
)
