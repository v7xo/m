//go:build test_unit

package losetup_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/losetup_thelper"
	hlp "rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func Test_NewDevice_Happy(t *testing.T) {
	ctx := hlp.Pre(t)

	devInfo := losetup_thelper.Create(t, ctx, model.DataSize{Unit: model.DataSizeUnitB, UnitAmount: 1024 * 1024 * 8})
	s := losetup_thelper.New(t, ctx)

	list, err := s.List(ctx)
	require.NoError(t, err)

	require.True(t, slices.Contains(list, devInfo), "created loop device not found among list of loop devices")
}
