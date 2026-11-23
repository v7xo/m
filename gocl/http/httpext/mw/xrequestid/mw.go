package xrequestid

import (
	"net/http"

	"google.golang.org/grpc/metadata"

	"rn.m.vo7.io/vo7io/m/gocl/appmeta/info"
	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	httpext "rn.m.vo7.io/vo7io/m/gocl/http/httpext"
	"rn.m.vo7.io/vo7io/m/gocl/http/httpext/mw"
	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/misc/uuidext"
)

func NewWithLog(log logext.Logger) mw.Func {
	if log == nil {
		log = logext.NopLogger
	}

	return func(next http.Handler) http.Handler {
		hnd := func(w http.ResponseWriter, r *http.Request) {
			enrichedReq, logAttrs, err := httpext.RequestWithAttrs(r, true)
			if err != nil {
				log.Err(r.Context(), err).Msg("failed to enrich HTTP request with logging attributes")

				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(placeholder.ErrDetailedMessageOmitted))

				return
			}

			r = enrichedReq

			reqID, err := uuidext.RequestIDHex()
			if err != nil {
				log.Err(r.Context(), err).
					Attrs(logAttrs.ToDict()).
					Str(attrname.LogScope, logScope).
					Str(attrname.Middleware, middlewareLogName).
					Msg("failed to enrich HTTP request with generated request ID, continue request handling")

				next.ServeHTTP(w, r)

				return
			}

			log.Info(r.Context()).
				Attrs(logAttrs.ToDict()).
				Str(attrname.LogScope, logScope).
				Str(attrname.Middleware, middlewareLogName).
				Str(attrname.RequestID, reqID).
				Msgf("use newly generated request ID '%s' for this HTTP request serving", reqID)

			ctx := r.Context()
			ctx = attrname.DictContext(ctx, attrname.Dict{attrname.RequestID: reqID})
			ctx = contextext.WithRequestIDReplace(ctx, reqID)
			ctx = metadata.AppendToOutgoingContext(ctx, string(attrname.XRequestID), reqID, string(attrname.XUserAgent), info.UserAgentShort())

			r = r.WithContext(ctx)

			w.Header().Set("X-Request-ID", reqID)
			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(hnd)
	}
}
