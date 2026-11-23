package cors

import (
	"fmt"
	"time"

	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

type Config struct {
	AllowedMethods     string
	AllowedHeaders     string
	AllowedOrigins     string
	ExposedHeaders     string
	MaxAge             time.Duration `validate:"min=1ns"`
	OptionsPassthrough bool
	AllowCredentials   bool
}

func (cfg *Config) BindFlagSetPrefix(prefix string, fls *pflag.FlagSet) (err error) {
	prefix = flenv.NormalizePrefix(prefix)

	fls.StringVar(
		&cfg.AllowedMethods,
		prefix+"allowed-methods",
		DefaultAllowedMethods,
		"comma separated list of methods the client allowed to use with cross domain requests")

	fls.StringVar(
		&cfg.AllowedOrigins,
		prefix+"allowed-origins",
		DefaultAllowedOrigins,
		"comma separated list of origins a cross-domain request can be executed from."+
			"\n\tif the special '*' value is present in the list, all origins will be allowed."+
			"\n\tan origin may contain a wildcard (*) to replace 0 or more characters (i.e.: http://*.domain.com)."+
			"\n\tusage of wildcards implies a small performance penalty."+
			"\n\tonly one wildcard can be used per origin")

	fls.StringVar(
		&cfg.AllowedHeaders,
		prefix+"allowed-headers",
		DefaultAllowedHeaders,
		"comma separated list of non simple headers the client is allowed to use with cross-domain requests."+
			"\n\tif the special '*' value is present in the list, all headers will be allowed")

	fls.DurationVar(
		&cfg.MaxAge,
		prefix+"max-age",
		DefaultMaxAge,
		"indicates how long the results of a preflight request can be cached")

	fls.BoolVar(
		&cfg.OptionsPassthrough,
		prefix+"options-passthrough",
		DefaultOptionsPassthrough,
		"instructs preflight to let other potential next handlers to process the OPTIONS method."+
			"\n\tturn this on if your application handles OPTIONS")

	fls.BoolVar(
		&cfg.AllowCredentials,
		prefix+"allow-credentials",
		DefaultAllowCredentials,
		"indicates whether the request can include user credentials like cookies, HTTP authentication or client side SSL certificates")

	return nil
}

func (cfg *Config) BindFlags(fls *pflag.FlagSet) error {
	return cfg.BindFlagSetPrefix("", fls)
}

func (cfg *Config) Validate(validator validation.Validator) error {
	return validator.Struct(cfg)
}

// DefaultConfig returns default CORS middleware config constructed from flags and environment vars.
func DefaultConfig() (Config, error) {
	var cfg Config

	if err := flenv.PopulatePrefix(ConfigFlagPrefix, &cfg); err != nil {
		return Config{}, fmt.Errorf("error while constructing default config of CORS middleware; err: %w", err)
	}

	return cfg, nil
}
