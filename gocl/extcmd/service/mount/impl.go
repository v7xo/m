package mount

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync/atomic"

	"rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/findmnt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/osutil"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type impl struct {
	fopt.LogOpt

	findmnt findmnt.Findmnt
}

func (s *impl) MountV2(ctx context.Context, params MountParams) (mi model.MountInfo, unmount func(context.Context) error, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to execute mount command using params '%s'; %w", stringsext.Stringer(params), err)

			s.Log().Err(ctx, err).Send()

			return
		}

		s.Log().Trace(ctx).Msgf("succeeded to execute mount command using params '%s', mounted '%s'", params, mi)
	}()

	switch paramsTyped := params.(type) {
	case MountParamsDeviceBase:
		_, mi, unmount, err = s.mountUsingDeviceBase(ctx, paramsTyped, true)

		return
	case nil:
		return model.MountInfo{}, nil, errors.New("params argument can't be nil")
	default:
		return model.MountInfo{}, nil, fmt.Errorf("params type '%T' not supported", params)
	}
}

func (s *impl) Mount(ctx context.Context, params MountParams) (target string, unmount func(context.Context) error, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to execute mount command using params '%s'; %w", stringsext.Stringer(params), err)

			s.Log().Err(ctx, err).Send()

			return
		}

		s.Log().Trace(ctx).Msgf("succeeded to execute mount command using params '%s', actual target (mount point) '%s'", params.String(), target)
	}()

	switch paramsTyped := params.(type) {
	case MountParamsDeviceBase:
		target, _, unmount, err = s.mountUsingDeviceBase(ctx, paramsTyped, false)

		return
	case nil:
		return "", nil, errors.New("params argument can't be nil")
	default:
		return "", nil, fmt.Errorf("params type '%T' not supported", params)
	}
}

func (s *impl) Umount(ctx context.Context, params UmountParams) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to execute unmount command using params '%s'; %w", stringsext.Stringer(params), err)

			s.Log().Err(ctx, err).Send()

			return
		}

		s.Log().Trace(ctx).Msgf("unmount command executed successfully using params '%s'", params.String())
	}()

	switch paramsTyped := params.(type) {
	case UnmountParamsTargetBase:
		return s.unmountUsingTargetBase(ctx, paramsTyped)
	case nil:
		return errors.New("params argument can't be nil")
	default:
		return fmt.Errorf("params type '%T' not supported", params)
	}
}

func (s *impl) unmountUsingTargetBase(ctx context.Context, p UnmountParamsTargetBase) error {
	if err := p.Validate(); err != nil {
		return err
	}

	cmdArgs := []string{"--lazy", p.Target}

	return osutil.Exec(
		ctx,
		osutil.ExecParamsNullOutput{
			Log:               s.Log(),
			Name:              unmountCommandName,
			RequiresSuperUser: true,

			Args: cmdArgs,
		})
}

func (s *impl) mountUsingDeviceBase(ctx context.Context, params MountParamsDeviceBase, retMountInfo bool) (_ string, mi model.MountInfo, unmount func(context.Context) error, err error) {
	if err = params.Validate(); err != nil {
		return "", model.MountInfo{}, nil, err
	}

	var (
		rmTarget       func() error
		rmTargetCalled atomic.Bool
	)

	if params.Target == "" {
		var tmpDir string

		if tmpDir, err = osutil.MkdirTemp("v7xo_gocl_mount_*"); err != nil {
			return "", model.MountInfo{}, nil, fmt.Errorf("failed to make temporary directory for mount point (for mount using params '%s'); %w", params, err)
		}

		s.Log().Trace(ctx).Msgf("mount params '%s' doesn't specify mount target, temporary directory '%s' will be used as target", params, tmpDir)

		rmTarget = func() error {
			if !rmTargetCalled.CompareAndSwap(false, true) {
				return nil
			}

			if remErr := os.Remove(tmpDir); remErr != nil {
				return fmt.Errorf("failed to remove temporary mount point directory '%s' (mount params '%s'); %w", tmpDir, params, remErr)
			}

			return nil
		}

		defer func() {
			if err == nil {
				return
			}

			err = errs.Join(err, rmTarget())
		}()

		params.Target = tmpDir
	}

	cmdArgs := []string{}

	if typeName, ok := params.Type.ToLinuxKernelKnownName(); ok {
		cmdArgs = append(cmdArgs, "-t", typeName)
	}

	if params.Options.Count() > 0 {
		cmdArgs = append(cmdArgs, "-o", params.Options.ToMountCommandString())
	}

	cmdArgs = append(cmdArgs, params.Source.Path(), params.Target)

	err = osutil.Exec(
		ctx,
		osutil.ExecParamsNullOutput{
			Log:               s.Log(),
			Name:              mountCommandName,
			RequiresSuperUser: true,

			Args: cmdArgs,
		})
	if err != nil {
		return "", model.MountInfo{}, nil, err
	}

	unmount = func(ctx context.Context) error {
		if unmountErr := s.Umount(ctx, UnmountParamsTargetBase{Target: params.Target}); unmountErr != nil || rmTarget == nil {
			return unmountErr
		}

		return rmTarget()
	}

	defer func() {
		if err != nil {
			err = errs.Join(err, unmount(contextext.AppGracefulEnd(ctx)))
		}
	}()

	switch {
	case params.TakeTargetOwnership:
		err = osutil.Exec(
			ctx,
			osutil.ExecParamsNullOutput{
				Log:               s.Log(),
				Name:              "chown",
				RequiresSuperUser: true,

				Args: []string{
					strconv.FormatInt(int64(os.Geteuid()), 10) + ":" + strconv.FormatInt(int64(os.Getegid()), 10),
					params.Target,
				},
			})
		if err != nil {
			return "", model.MountInfo{}, nil, fmt.Errorf("failed to take ownership '%s'; %w", params.Target, err)
		}
	}

	if retMountInfo {
		var found bool

		mi, found, err = s.findmnt.ByTarget(ctx, params.Target)
		if err != nil {
			return "", model.MountInfo{}, nil, err
		}

		if !found {
			return "",
				model.MountInfo{},
				nil,
				fmt.Errorf("unexpected result of searching the mounted file system by target '%s', search result doesn't contain expected data", params.Target)
		}
	}

	return params.Target, mi, unmount, nil
}

func New(opts ...OptionFn) (res Mount, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", stringsext.TypeStringFor[impl](), err)
		}
	}()

	impl := &impl{}
	if err = fopt.Apply(impl, opts...); err != nil {
		return nil, err
	}

	if impl.findmnt == nil {
		return nil, fmt.Errorf("no service '%s' specified", stringsext.TypeStringForExpr(impl.findmnt))
	}

	return impl, nil
}
