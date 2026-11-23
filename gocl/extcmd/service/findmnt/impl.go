package findmnt

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	extcmd_model "rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/osutil"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type impl struct{ fopt.LogOpt }

func (s *impl) List(ctx context.Context) (extcmd_model.MountInfoList, error) {
	return s.listRealAndUniq(ctx, "")
}

func (s *impl) ByPath(ctx context.Context, path string) (_ extcmd_model.MountInfo, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to detemine mounted file system for specified path '%s'; %w", path, err)
		}
	}()

	if path == "" {
		return extcmd_model.MountInfo{}, errors.New("path can't be an empty string")
	}

	list, err := s.listRealAndUniq(ctx, path)
	if err != nil {
		return extcmd_model.MountInfo{}, err
	}

	if len(list) == 0 {
		return extcmd_model.MountInfo{}, errs.ErrNotFound
	} else if len(list) != 1 {
		return extcmd_model.MountInfo{}, fmt.Errorf("%w; more than one mounted file systems found", errs.ErrAmbiguousMatch)
	}

	return list[0], nil
}

func (s *impl) ByTarget(ctx context.Context, target string) (extcmd_model.MountInfo, bool, error) {
	if target == "" {
		return extcmd_model.MountInfo{}, false, errors.New("mount target can't be an empty string to find mount")
	}

	list, err := s.listRealAndUniq(ctx, target)
	if err != nil {
		return extcmd_model.MountInfo{}, false, err
	}

	list = list.All(func(mi extcmd_model.MountInfo) bool { return mi.Target == target })

	if len(list) == 0 {
		return extcmd_model.MountInfo{}, false, nil
	} else if len(list) != 1 {
		return extcmd_model.MountInfo{}, false, fmt.Errorf("%w; more than one mounted file systems found by specified mount target '%s'", errs.ErrAmbiguousMatch, target)
	}

	return list[0], true, nil
}

func (s *impl) BySourceUUIDList(ctx context.Context, source uuid.UUID) (extcmd_model.MountInfoList, error) {
	if source == uuid.Nil {
		return nil, fmt.Errorf("mount source UUID argument can't be nil ('%s') to find mounts", uuid.Nil)
	}

	list, err := s.listRealAndUniq(ctx, "")
	if err != nil {
		return nil, err
	}

	return list.All(func(mi extcmd_model.MountInfo) bool { return mi.FileSystemUUID == source }), nil
}

func (s *impl) MountOptionValueByName(ctx context.Context, mountPoint string, name string) (res string, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to get mount option value for option name '%s' and mount point '%s'; %w", name, mountPoint, err)
		}
	}()

	if name == "" {
		err = errs.BadArgsWrapWith(errors.New("option name can't be an empty string"))

		return
	}

	optsString, err := s.MountOptionsString(ctx, mountPoint)
	if err != nil {
		return
	}

	opts := strings.Split(optsString, ",")
	optSearchPrefix := name + "="

	for i := range opts {
		if strings.HasPrefix(opts[i], optSearchPrefix) {
			res = opts[i][len(optSearchPrefix):]

			return
		}
	}

	err = errors.New("no such mount option")

	return
}

func (s *impl) MountOptionsString(ctx context.Context, mountPoint string) (res string, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to get mount options string for mount point '%s'; %w", mountPoint, err)
		}
	}()

	output, err := osutil.ExecTextOutput(
		ctx,
		osutil.ExecParamsTextOutput{
			Log:            s.Log(),
			Name:           findmntCmd,
			Args:           []string{"-n", "-o", "OPTIONS", mountPoint},
			StdoutMinLines: 1,
			StdoutMaxLines: 1,
		})
	if err != nil {
		return
	}

	res = string(output[0])

	return
}

func (s *impl) RootFSTypeCode(ctx context.Context) (res extcmd_model.FileSystemTypeCode, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to determine root file system type; %w", err)
		}
	}()

	output, err := osutil.ExecTextOutput(
		ctx,
		osutil.ExecParamsTextOutput{
			Log:            s.Log(),
			Name:           findmntCmd,
			Args:           []string{"-n", "-M", "/", "-o", "FSTYPE"},
			StdoutMinLines: 1,
			StdoutMaxLines: 1,
		})
	if err != nil {
		return
	}

	err = res.FromString(string(output[0]))

	return
}

func (s *impl) listRealAndUniq(ctx context.Context, targetCmdArg string) (res extcmd_model.MountInfoList, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to list mounted file systems (--target '%s'); %w", targetCmdArg, err)

			return
		}

		s.Log().
			Info(ctx).
			Msgf("got list of mounted file systems (unique and real, --target '%s'), list contains '%d' items", targetCmdArg, len(res))
	}()

	args := []string{
		"--json",
		"--nofsroot",
		"--real",
		"-b",
		"--uniq",
	}

	if len(targetCmdArg) != 0 {
		args = append(args, "--target", targetCmdArg)
	}

	args = append(args, "-o", "+UUID,LABEL,SIZE,AVAIL,FSROOT")

	outputReader, err := osutil.ExecTextOutputReader(
		ctx,
		osutil.ExecParamsTextOutput{
			Args:           args,
			Log:            s.Log(),
			StdoutMaxLines: maxCmdOutputBuffer,
			Name:           findmntCmd,
		})
	if err != nil {
		return
	}

	err = res.FromFindmntOutputReaderJSON(outputReader)

	return
}

func New(opts ...OptionFn) (_ Findmnt, err error) {
	// TODO: add Go build tag to include this func only for Unix OS.
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", stringsext.TypeStringFor[impl](), err)
		}
	}()

	impl := &impl{}
	if err = fopt.Apply(impl, opts...); err != nil {
		return nil, err
	}

	return impl, nil
}
