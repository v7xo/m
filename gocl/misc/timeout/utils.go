package timeout

import (
	"fmt"
	"math"
	"math/rand"
	randv2 "math/rand/v2"
	"time"
)

// Jittered generate variation of time delay using specified base value and jitter.
func Jittered(base time.Duration, jitter float64) (time.Duration, error) {
	return JitteredSeed(base, jitter, 0)
}

// JitteredSeed generate variation of time delay using specified base value and jitter.
func JitteredSeed(base time.Duration, jitter float64, seed int64) (res time.Duration, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to generate jittered delay value; %w", err)
		}
	}()

	if base < 0 {
		return 0, fmt.Errorf("specified value '%s' of base interval is not valid, can't be negative", base)
	} else if math.IsNaN(jitter) || math.IsInf(jitter, 0) {
		return 0, fmt.Errorf("specified value '%v' of jitter is not valid, can't be either NaN or Inf", jitter)
	} else if jitter < 0 {
		return 0, fmt.Errorf("specified value '%v' of jitter is not valid, can't be negative", jitter)
	} else if jitter == 0 {
		jitter = DefaultJitter
	}

	var mul float64

	if seed == 0 {
		mul = randv2.Float64() // nolint:gosec
	} else {
		mul = rand.New(rand.NewSource(seed)).Float64() // nolint:gosec
	}

	j := jitter * (mul*2.0 - 1.0)

	res = max(base+time.Duration(float64(base)*j), 0)

	return
}

func MustJittered(base time.Duration, jitter float64) time.Duration {
	res, err := Jittered(base, jitter)
	if err != nil {
		panic(err)
	}

	return res
}

func MustJitteredDefault(base time.Duration) time.Duration {
	res, err := Jittered(base, DefaultJitter)
	if err != nil {
		panic(err)
	}

	return res
}
