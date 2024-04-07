package backend

import (
	"net/http"

	"github.com/ksysoev/wasabi"
)

type RequestFactory func(req wasabi.Request) (*http.Request, error)
