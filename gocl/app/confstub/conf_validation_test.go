//go:build test_unit

package confstub_test

import (
	"testing"

	"rn.m.vo7.io/vo7io/m/gocl/app/confstub"
	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/diag/prompusher"
	"rn.m.vo7.io/vo7io/m/gocl/status/level"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper/stdcase"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type TestConfig struct {
	confstub.Log
	confstub.DiagEndpoint
}

func (c TestConfig) Validate(v validation.Validator) error {
	return v.Struct(c)
}

func Test_ConfValidation(t *testing.T) {
	thelper.Pre(t)

	type testCase struct {
		inputConfig    flenv.Conf
		inputValidator validation.Validator

		stdcase.CommonProps
	}

	var (
		defValidator  = validation.DefaultValidator()
		defPusherConf = prompusher.Conf{
			GatewayAddress: prompusher.DefaultConf().GatewayAddress,
			Interval:       prompusher.DefaultConf().Interval,
			Timeout:        prompusher.DefaultConf().Timeout,
		}
	)

	testCases := []testCase{
		{
			CommonProps: stdcase.CommonProps{
				Desc: "test000: invalid diag endpoint listen port",

				FailureProps: stdcase.FailureProps{
					ShouldFail:      true,
					ShouldFailAsErr: &validation.Error{},
				},
			},

			inputConfig: TestConfig{
				Log: confstub.Log{
					LibString:    string(logext.LibProviderGoKit),
					FormatString: string(logext.FormatJSON),
					LevelString:  string(level.Info),
				},
				DiagEndpoint: confstub.DiagEndpoint{
					ListenPort:    0,
					MetricsPusher: defPusherConf,
				},
			},
			inputValidator: defValidator,
		},
		{
			CommonProps: stdcase.CommonProps{
				Desc: "test001: invalid diag endpoint listen port",

				FailureProps: stdcase.FailureProps{
					ShouldFail:      true,
					ShouldFailAsErr: &validation.Error{},
				},
			},

			inputConfig: TestConfig{
				Log: confstub.Log{
					LibString:    string(logext.LibProviderGoKit),
					FormatString: string(logext.FormatJSON),
					LevelString:  string(level.Info),
				},
				DiagEndpoint: confstub.DiagEndpoint{
					ListenPort:    65536,
					MetricsPusher: defPusherConf,
				},
			},
			inputValidator: defValidator,
		},
		{
			CommonProps: stdcase.CommonProps{
				Desc: "test002: valid diag endpoint listen port",
			},

			inputConfig: TestConfig{
				Log: confstub.Log{
					LibString:    string(logext.LibProviderGoKit),
					FormatString: string(logext.FormatJSON),
					LevelString:  string(level.Info),
				},
				DiagEndpoint: confstub.DiagEndpoint{
					ListenPort:          65535,
					MetricsPathPrefix:   "metrics",
					MetricsPusher:       defPusherConf,
					PProfPathPrefix:     "debug/pprof",
					LivenessPathPrefix:  "/liveness",
					ReadinessPathPrefix: "/readiness",
					ListenHost:          "0.0.0.0",
				},
			},
			inputValidator: defValidator,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(
			tc.Desc,
			func(t *testing.T) {
				tc.PreLog(t)

				err := tc.inputConfig.Validate(tc.inputValidator)
				tc.RequireConformError(t, err)
			},
		)
	}
}
