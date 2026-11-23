package cors

import (
	"fmt"

	"github.com/go-chi/cors"

	"rn.m.vo7.io/vo7io/m/gocl/http/httpext/mw"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

type middleware struct {
	cors.Options
}

func New(opts ...OptionFn) (m mw.Func, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error while creating instance of CORS HTTP middleware; err: %w", err)
		}
	}()

	middleware := &middleware{}

	if err = fopt.Apply(middleware, opts...); err != nil {
		return
	}

	m = cors.Handler(middleware.Options)

	return
}
