package backoff

import (
	"time"

	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/validation"
)

// ExponentialConfig defines the configuration options for exponential backoff.
type ExponentialConfig struct {
	// BaseDelay is the amount of time to backoff after the first failure.
	BaseDelay time.Duration `validate:"min=0s"`

	// Multiplier is the factor with which to multiply backoffs after a
	// failed retry. Should ideally be greater than 1.
	Multiplier float64 `validate:"gte=1.0"`

	// Jitter is the factor with which backoffs are randomized.
	Jitter float64 `validate:"gte=0.1"`

	// MaxDelay is the upper bound of backoff delay.
	MaxDelay time.Duration `validate:"min=1s"`
}

func (cfg *ExponentialConfig) BindFlagSetPrefix(prefix string, fls *pflag.FlagSet) error {
	prefix = flenv.NormalizePrefix(prefix)

	fls.DurationVar(
		&cfg.BaseDelay,
		prefix+"base-delay",
		DefaultExpBackoffBaseDelay,
		"amount of time to backoff after the first failure")
	fls.DurationVar(
		&cfg.MaxDelay,
		prefix+"max-delay",
		DefaultExpBackoffMaxDelay,
		"upper bound of backoff delay")
	fls.Float64Var(
		&cfg.Multiplier,
		prefix+"multiplier",
		DefaultExpBackoffMultiplier,
		"factor with which to multiply backoffs after a failed retry")
	fls.Float64Var(
		&cfg.Jitter,
		prefix+"jitter",
		DefaultExpBackoffJitter,
		"factor with which backoffs are randomized")

	return nil
}

func (cfg *ExponentialConfig) BindFlags(fls *pflag.FlagSet) error {
	return cfg.BindFlagSetPrefix("", fls)
}

func (cfg *ExponentialConfig) Validate(v validation.Validator) error { return v.Struct(cfg) }
