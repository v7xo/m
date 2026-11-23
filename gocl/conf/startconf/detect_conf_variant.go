package startconf

import (
	"errors"
	"fmt"
	"os"
)

func DetectConfVariant() (_ ConfVariant, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to detect configuration variant; %w", err)
		}
	}()

	m := map[string]ConfVariant{
		DevelopLocalConfYamlFileName: ConfVariantDevelopLocal,
		DevelopConfYamlFileName:      ConfVariantDevelop,
	}

	for fileName, variant := range m {
		if _, err = os.Stat(fileName); os.IsNotExist(err) || errors.Is(err, os.ErrNotExist) {
			err = nil

			continue
		} else if err != nil {
			return "", err
		}

		return variant, nil
	}

	return ConfVariantIgnore, nil
}
