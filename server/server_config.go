package server

import (
	"time"
)

type ServerConfig struct {
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
}

var DefaultServerConfig = ServerConfig{
	ReadHeaderTimeout: ReadHeaderTimeoutSeconds * time.Second,
	ReadTimeout:       ReadTimeoutSeconds * time.Second,
}

const ReadHeaderTimeoutSeconds = 3

const ReadTimeoutSeconds = 30
