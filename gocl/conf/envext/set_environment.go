package envext

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"rn.m.vo7.io/vo7io/m/gocl/conf/startconf"
)

// [SetEnvironment2] sets environment accordingly with the specified [startconf.Conf].
//
// Consider, useConfVariant and useUserProfilerConfSubdir overrides specified configuration.
func SetEnvironment2(
	useConfVariant startconf.ConfVariant,
	useUserProfileConfSubdir string,
	conf startconf.Conf,
) (_ bool, err error) {
	variant := useConfVariant
	if variant.IsUnspecified() {
		variant = conf.ConfVariant()
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to set environment; %w", err)
		}
	}()

	if variant.IsIgnore() {
		return false, nil
	} else if variant.IsAuto() {
		return false, fmt.Errorf("invalid value of configuration variant, can't be '%s'", startconf.ConfVariantAuto)
	}

	var (
		allPreceding  = conf.PrecedingFileNames()
		allSucceeding = conf.SucceedingFileNames()
		entryDir      string
	)

	switch variant {
	case startconf.ConfVariantDevelop:
		allPreceding = append(allPreceding, startconf.DevelopConfYamlFileName)
		entryDir = startconf.DevelopConfEntryDir
	case startconf.ConfVariantDevelopLocal:
		allPreceding = append(allPreceding, startconf.DevelopLocalConfYamlFileName, startconf.DevelopConfYamlFileName)
		entryDir = startconf.DevelopConfEntryDir
	case startconf.ConfVariantTest:
		allPreceding = append(allPreceding, startconf.TestConfYamlFileName)
		entryDir = startconf.TestConfEntryDir
	case startconf.ConfVariantTestLocal:
		allPreceding = append(allPreceding, startconf.TestLocalConfYamlFileName, startconf.TestConfYamlFileName)
		entryDir = startconf.TestConfEntryDir
	case startconf.ConfVariantUserProfile:
		if entryDir, allPreceding, allSucceeding, err = UserProfileConfDir(useUserProfileConfSubdir, conf); err != nil {
			return false, err
		}
	case startconf.ConfVariantAuto, startconf.ConfVariantUnspecified, startconf.ConfVariantIgnore:
		fallthrough
	default:
		return false, fmt.Errorf("invalid/unknown configuration variant '%s'", variant)
	}

	return AddenvFromProvider(context.Background(), VarProviderFile(entryDir, append(allPreceding, allSucceeding...)...))
}

// [SetEnvironment] sets environment accordingly with the specified variant (see [ConfVariant]).
//
// Deprecated: use [SetEnvironment2] instead.
func SetEnvironment(confVariant startconf.ConfVariant, userProfileConfSubdir string, precedingFileNames, succeedingFileNames []string) (_ bool, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to set environment; %w", err)
		}
	}()

	if confVariant.IsIgnore() {
		return false, nil
	} else if confVariant.IsAuto() {
		return false, fmt.Errorf("invalid value of configuration variant, can't be '%s'", startconf.ConfVariantAuto)
	}

	preFileNames := make([]string, 0, 2+len(precedingFileNames)+len(succeedingFileNames)) //nolint:mnd
	entryDir := ""

	preFileNames = append(preFileNames, precedingFileNames...)

	switch confVariant {
	case startconf.ConfVariantDevelop:
		preFileNames = append(preFileNames, startconf.DevelopConfYamlFileName)
		entryDir = startconf.DevelopConfEntryDir
	case startconf.ConfVariantDevelopLocal:
		preFileNames = append(preFileNames, startconf.DevelopLocalConfYamlFileName, startconf.DevelopConfYamlFileName)
		entryDir = startconf.DevelopConfEntryDir
	case startconf.ConfVariantTest:
		preFileNames = append(preFileNames, startconf.TestConfYamlFileName)
		entryDir = startconf.TestConfEntryDir
	case startconf.ConfVariantTestLocal:
		preFileNames = append(preFileNames, startconf.TestLocalConfYamlFileName, startconf.TestConfYamlFileName)
		entryDir = startconf.TestConfEntryDir
	case startconf.ConfVariantAuto, startconf.ConfVariantUnspecified, startconf.ConfVariantIgnore, startconf.ConfVariantUserProfile:
		fallthrough
	default:
		return false, fmt.Errorf("invalid/unsupported configuration variant '%s'", confVariant)
	}

	return AddenvFromProvider(context.Background(), VarProviderFile(entryDir, append(preFileNames, succeedingFileNames...)...))
}

func UserProfileConfDir(useConfSubdir string, conf startconf.Conf) (dir string, preceding, succeeding []string, err error) {
	switch {
	case useConfSubdir == "":
		confDir := conf.UserProfileConfDir()
		if confDir == "" {
			return "", nil, nil, errors.New("no user profile configuration subdirectory specified")
		}

		return confDir, append(conf.PrecedingFileNames(), UserProfileConfYamlFileName), conf.SucceedingFileNames(), nil
	}

	if filepath.IsAbs(useConfSubdir) {
		return "", nil, nil, errors.New("user profile configuration subdirectory '" + useConfSubdir + "' is not valid; it can't be an absolute path")
	}

	osUserConfsDir, osDirErr := os.UserConfigDir()
	if osDirErr != nil {
		return "", nil, nil, fmt.Errorf("failed to determine OS user configurations directory; %w", osDirErr)
	}

	return filepath.Join(osUserConfsDir, useConfSubdir), append(conf.PrecedingFileNames(), UserProfileConfYamlFileName), conf.SucceedingFileNames(), nil
}
