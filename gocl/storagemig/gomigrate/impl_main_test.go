//go:build test_postgres

package gomigrate_test

import (
	"os"
	"testing"

	"rn.m.vo7.io/vo7io/m/gocl/storagemig/util"
)

func TestMain(m *testing.M) {
	util.SetAppReleaseFn(func() string { return appReleaseFn })

	os.Exit(m.Run())
}
