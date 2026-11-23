package server

import (
	"context"
	"net/http"
)

type DriverFn func(stop context.Context, s *http.Server) error
