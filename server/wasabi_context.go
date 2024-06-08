package server

import (
	"context"
	"time"
)

type WasabiContext interface {
	context.Context
	GetHttpConfig() HttpConfig
}

type WasabiDefaultContextImpl struct {
	context.Context
}

func (w WasabiDefaultContextImpl) GetHttpConfig() HttpConfig {
	return *getDefaultHttpConfig()
}

type HttpConfig struct {
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
}

func NewWasabiDefaultContext(fromContext context.Context) WasabiDefaultContextImpl {
	var ctx = new(WasabiDefaultContextImpl)
	return *ctx
}

var defaultHttpConfig *HttpConfig

func getDefaultHttpConfig() *HttpConfig {
	if defaultHttpConfig == nil {
		defaultHttpConfig = &HttpConfig{
			ReadHeaderTimeout: 3 * time.Second,
			ReadTimeout:       30 * time.Second,
		}
	}

	return defaultHttpConfig
}
