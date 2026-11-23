package envext

import (
	"fmt"
	"os"
	"strconv"
)

// RunningCIContainer returns true if runtime is in CI container.
//
// Function relies on presence either environment 'CI' or 'RUNNING_IN_APP_BUILDER_CONTAINER'.
func RunningCIContainer() bool {
	checkEnv := func(n string) bool {
		val, exists := os.LookupEnv(n)
		if !exists {
			return false
		}

		valBool, err := strconv.ParseBool(val)
		if err != nil {
			return false
		}

		return valBool
	}

	return checkEnv("CI") || checkEnv("RUNNING_IN_APP_BUILDER_CONTAINER")
}

// DebugTest returns true if test debug option enabled (func relies on presence and value of DEBUG_TEST environment variable).
func DebugTest() bool {
	val, exists := os.LookupEnv(DebugTestEnvVar)
	if !exists {
		return false
	}

	valBool, err := strconv.ParseBool(val)
	if err != nil {
		return false
	}

	return valBool
}

// SetenvIfNotExists sets environment var if one not exists.
func SetenvIfNotExists(name string, val string) (actualVal string, set bool, err error) {
	actualVal, ok := os.LookupEnv(name)
	if !ok {
		if err := os.Setenv(name, val); err != nil {
			return "", false, fmt.Errorf("error while setting environment variable; variable: %s; err: %w", name, err)
		}

		set = true
		actualVal = val
	}

	return
}

func TestCreateStorageUser() bool {
	val, exists := os.LookupEnv("TEST_CREATE_STORAGE_USER")
	if !exists {
		return false
	}

	valBool, err := strconv.ParseBool(val)
	if err != nil {
		return false
	}

	return valBool
}
