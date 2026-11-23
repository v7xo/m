package hashauth

import "context"

type Verifier interface {
	Verify(ctx context.Context, secret string) error
}
