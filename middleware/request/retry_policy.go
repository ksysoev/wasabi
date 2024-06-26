package request

import (
	"time"

	"github.com/ksysoev/wasabi"
)

func LinearRetryPolicy(maxRetries int, interval time.Duration, next wasabi.RequestHandler, conn wasabi.Connection, req wasabi.Request) error {
	var err error

	ticker := time.NewTicker(interval)

	defer ticker.Stop()

	for i := 0; i < maxRetries; i++ {
		err = next.Handle(conn, req)
		if err == nil {
			return nil
		}

		ticker.Reset(interval)

		select {
		case <-req.Context().Done():
			return req.Context().Err()
		case <-ticker.C:
		}
	}

	return err
}
