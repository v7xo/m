package recovery

import (
	"errors"
	"net/http"

	attrname "rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	clhttp "rn.m.vo7.io/vo7io/m/gocl/http/httpext"
	"rn.m.vo7.io/vo7io/m/gocl/http/httpext/mw"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

func NewWithLog(lg logext.Logger, preventErrDetailedMessage bool) mw.Func {
	if lg == nil {
		lg = logext.NopLogger
	}

	fnRes := func(next http.Handler) http.Handler {
		hnd := func(w http.ResponseWriter, r *http.Request) {
			panicking := true

			defer func() {
				if !panicking {
					return
				}

				errToLog := errs.PanicErrUnconditionally(recover())
				if !errors.Is(errToLog, http.ErrAbortHandler) {
					_, logAttrs, logAttrsErr := clhttp.RequestWithAttrs(r, false)
					if logAttrsErr != nil {
						errToLog = errors.Join(errToLog, logAttrsErr)

						reqURLString := placeholder.Nil
						if r.URL != nil {
							reqURLString = r.URL.Redacted()
						}

						logAttrs = clhttp.RequestAttrs(
							attrname.Dict{
								attrname.HTTPReqURL:    reqURLString,
								attrname.HTTPReqMethod: r.Method,
								attrname.UserAgent:     r.UserAgent(),
							},
						)
					}

					lg.Err(r.Context(), errToLog).
						Attrs(logAttrs.ToDict()).
						Str(attrname.LogScope, logScope).
						Str(attrname.Middleware, middlewareLogName).
						Msg("panic occurred; see error details")

					w.WriteHeader(http.StatusInternalServerError)

					if preventErrDetailedMessage {
						_, _ = w.Write([]byte(placeholder.ErrDetailedMessageOmitted))
					} else {
						_, _ = w.Write([]byte(errToLog.Error()))
					}
				}
			}()

			next.ServeHTTP(w, r)

			panicking = false
		}

		return http.HandlerFunc(hnd)
	}

	return fnRes
}
