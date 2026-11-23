package envext

import (
	"context"
	"fmt"
	"os"
)

func Addenv(vars map[string]string) (bool, error) {
	hasChanges := false

	for varName, varVal := range vars {
		if _, ok := os.LookupEnv(varName); ok {
			continue
		}

		if err := os.Setenv(varName, varVal); err != nil {
			return hasChanges, fmt.Errorf("failed to set environment '%s'; %w", varName, err)
		}

		hasChanges = true
	}

	return hasChanges, nil
}

func AddenvFromProvider(ctx context.Context, p VarProvider) (bool, error) {
	dict, err := p.VariablesDict(ctx)
	if err != nil {
		return false, err
	}

	return Addenv(dict)
}
