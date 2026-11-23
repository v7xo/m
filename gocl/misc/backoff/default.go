package backoff

func Default() Strategy { return &Exponential{Conf: DefaultConfig()} }

func DefaultConfig() ExponentialConfig {
	return ExponentialConfig{
		BaseDelay:  DefaultExpBackoffBaseDelay,
		Jitter:     DefaultExpBackoffJitter,
		MaxDelay:   DefaultExpBackoffMaxDelay,
		Multiplier: DefaultExpBackoffMultiplier,
	}
}
