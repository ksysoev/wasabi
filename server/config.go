package server

import (
	"time"
)

type Config struct {
	ReadHeaderTimeout time.Duration
	ReadTimeout       time.Duration
}

var DefaultConfig = Config{
	ReadHeaderTimeout: ReadHeaderTimeoutSeconds * time.Second,
	ReadTimeout:       ReadTimeoutSeconds * time.Second,
}

const ReadHeaderTimeoutSeconds = 3

const ReadTimeoutSeconds = 30
