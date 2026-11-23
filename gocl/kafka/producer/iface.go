package producer

import (
	"context"

	"rn.m.vo7.io/vo7io/m/gocl/kafka/header"
)

type Service interface {
	WriteMessage(ctx context.Context, topic string, key, val []byte, headers map[header.Name]string) error

	Close() error
}
