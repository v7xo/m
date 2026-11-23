package endpointfactory

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/endpoint"
	"rn.m.vo7.io/vo7io/m/gocl/diag/endpoint/http"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/diag/readiness"
	"rn.m.vo7.io/vo7io/m/gocl/misc/ptrof"
)

func Default() Fn {
	return func(confGetter endpoint.ConfGetter, l logext.Logger) (endpoint.Service, error) {
		var (
			conf *endpoint.Conf
			err  error
		)

		for {
			switch confGetter {
			case nil:
				if conf, err = ptrof.OfErr(endpoint.FlenvRootConf()); err != nil {
					return nil, err
				}
			default:
				if conf, err = confGetter.DiagnosticsEndpointConf(); err != nil {
					return nil, err
				} else if conf == nil {
					confGetter = nil

					continue
				}
			}

			break
		}

		return http.New(
			http.Option().
				ReadinessInspection(readiness.Registry()).
				Conf(*conf).
				Log(l),
		)
	}
}
