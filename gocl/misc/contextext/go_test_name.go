package contextext

import (
	"context"
)

func TestName(ctx context.Context) (string, bool) {
	if ctx == nil {
		return "", false
	}

	name, ok := ctx.Value(KeyGoTestName).(string)

	return name, ok
}
