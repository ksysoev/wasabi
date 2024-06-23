package server

import (
	"time"
)

type ServerConfig struct {
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
}

var DefaultServerConfig = ServerConfig{
	ReadHeaderTimeout: 3 * time.Second,
	ReadTimeout:       30 * time.Second,
}
