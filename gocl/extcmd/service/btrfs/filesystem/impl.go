package filesystem

import (
	"context"
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/mount"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/osutil"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/typestring"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type impl struct {
	fopt.LogOpt

	mount mount.Mount
}

func (s *impl) Mount(ctx context.Context, source model.BlockDeviceID, explicitTarget string, opts model.MountOptions) (target string, unmount func(context.Context) error, err error) {
	mi, unmount, err := s.MountV2(ctx, source, explicitTarget, opts)
	if err != nil {
		return "", nil, err
	}

	return mi.Target, unmount, nil
}

func (s *impl) MountV2(ctx context.Context, source model.BlockDeviceID, explicitTarget string, opts model.MountOptions) (mi model.MountInfo, unmount func(context.Context) error, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to mount file system of '%s' as BTRFS file system; %w", source, err)

			s.Log().Err(ctx, err).Send()

			return
		}

		s.Log().Trace(ctx).Msgf("succeeded to mount BTRFS file system of '%s' on '%s'", source, mi.Target)
	}()

	mountParams := mount.MountParamsDeviceBase{
		Options: opts,
		Source:  source,
		Target:  explicitTarget,
		Type:    model.FileSystemTypeCodeBtrfs,
	}

	mi, unmount, err = s.mount.MountV2(ctx, mountParams)

	return
}

func (s *impl) Make(ctx context.Context, params MakeParams) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to make BTRFS file system using specified params '%s'; %w", params, err)

			return
		}

		s.Log().Trace(ctx).Msgf("BTRFS file system successfully created on device(s) '%s'", params.BackingDevice)
	}()

	if err = params.Validate(); err != nil {
		return
	}

	var csumString string

	switch params.CSum {
	case model.BtrfsCSumBlake2:
		csumString = "blake2"
	case model.BtrfsCSumSHA256:
		csumString = "sha256"
	case model.BtrfsCSumCRC32C:
		csumString = "crc32c"
	case model.BtrfsCSumXXHash:
		csumString = "xxhash"
	case model.BtrfsCSumUnspecified:
		fallthrough
	default:
		return fmt.Errorf("check sum algorithm '%s' not supported", params.CSum)
	}

	args := make([]string, 0, 9+len(params.BackingDevice)) //nolint:mnd,revive

	if params.Force {
		args = append(args, "-f")
	}

	args = append(
		args,
		"-U", params.UUID.String(),
		"-L", params.Label,
		"--csum", csumString)

	if params.BlockGroupTree {
		args = append(
			args,
			"--features", "block-group-tree")
	}

	for i := range params.BackingDevice {
		args = append(args, params.BackingDevice[i].Path())
	}

	_, err = osutil.ExecTextOutputReader(
		ctx,
		osutil.ExecParamsTextOutput{
			Args:              args,
			Log:               s.Log(),
			StdoutMaxLines:    model.DefaultCommandOutputLines,
			Name:              "mkfs.btrfs",
			RequiresSuperUser: true,
		})

	return
}

func (s *impl) Info(ctx context.Context, d model.BlockDeviceID) (res model.BtrfsInfo, err error) {
	return res, errs.ErrNotSupported
}

func New(opts ...OptionFn) (_ FileSystem, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", typestring.TypeStringFor[impl](), err)
		}
	}()

	impl := &impl{}
	if err = fopt.Apply(impl, opts...); err != nil {
		return
	}

	if impl.mount == nil {
		return nil, fmt.Errorf("no service '%s' specified", typestring.TypeStringForExpr(impl.mount))
	}

	return impl, nil
}
