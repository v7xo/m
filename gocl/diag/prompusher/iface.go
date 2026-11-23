package prompusher

import (
	"context"
	"io"
)

type Pusher interface {
	io.Closer

	Start(stop context.Context) error
}
