package confset

import (
	"reflect"
	"slices"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/diag/endpoint"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
)

type (
	builder struct {
		userConfs               []userConf
		logConf                 bool
		diagnosticsEndpointConf bool
	}
	userConf struct {
		typ  reflect.Type
		conf flenv.Conf
	}
)

func (b *builder) Log(val bool) *builder { b.logConf = val; return b }

func (b *builder) DiagnosticsEndpoint(val bool) *builder {
	b.diagnosticsEndpointConf = val

	return b
}

func (b *builder) Conf[T flenv.Conf](conf T) *builder {
	b.userConfs = append(b.userConfs, userConf{typ: reflect.TypeFor[T](), conf: conf})

	return b
}

func (b *builder) Build() Set {
	set := &set{confs: map[reflect.Type]flenv.Conf{}}

	for _, item := range slices.Backward(b.userConfs) {
		if _, ok := set.confs[item.typ]; ok {
			continue
		}

		set.confs[item.typ] = item.conf
	}

	if _, ok := set.confs[logConfTyp]; !ok && b.logConf {
		set.confs[logConfTyp] = &logext.Conf{}
	}

	if _, ok := set.confs[diagnosticsEndpointConfTyp]; !ok && b.diagnosticsEndpointConf {
		set.confs[diagnosticsEndpointConfTyp] = &endpoint.Conf{}
	}

	return set
}

func Builder() *builder { return &builder{diagnosticsEndpointConf: true, logConf: true} }
