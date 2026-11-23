package losetup

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	modl "rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/osutil"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/typestring"
)

type impl struct {
	fopt.LogOpt
}

func (s *impl) Detach(ctx context.Context, n string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to detach a loop device '%s'; err: %w", n, err)

			s.Log().Err(ctx, err).Send()

			return
		}
	}()

	if len(n) == 0 {
		return errors.New("empty device name specified")
	}

	_, err = osutil.ExecTextOutput(
		ctx,
		osutil.ExecParamsTextOutput{
			Log:               s.Log(),
			Name:              "losetup",
			Args:              []string{"-d", n},
			StdoutMaxLines:    1,
			RequiresSuperUser: true,
		})

	return
}

func (s *impl) New(ctx context.Context, p NewParams) (res modl.LoopDeviceInfo, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new loop device using parameters '%s'; err: %w", p, err)

			s.Log().Err(ctx, err).Send()

			return
		}

		s.Log().
			Info(ctx).
			Msgf("new loop device '%s' created using params '%s' and attached", res, p)
	}()

	switch params := p.(type) {
	case nil:
		return res, errors.New("nil parameters")
	case NewParamsFile:
		return s.attachUsingRegularFileParams(ctx, &params, nil)
	default:
		return res, fmt.Errorf("unknown parameters type '%T'", p)
	}
}

func (s *impl) Attach(ctx context.Context, p AttachParams) (res modl.LoopDeviceInfo, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to attach a loop device using specified parameters '%s'; err: %w", p, err)

			s.Log().Err(ctx, err).Send()

			return
		}

		s.Log().Info(ctx).Msgf("loop device '%s' attached using parameters '%s'", res, p)
	}()

	switch params := p.(type) {
	case nil:
		return res, errors.New("nil parameters")
	case AttachParamsFile:
		return s.attachUsingRegularFileParams(ctx, nil, &params)
	default:
		return res, fmt.Errorf("unknown parameters type '%T'", p)
	}
}

func (s *impl) List(ctx context.Context) (res modl.LoopDeviceInfoList, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to list loop devices using 'losetup' command; err: %w", err)
		}
	}()

	output, err := osutil.ExecTextOutput(
		ctx,
		osutil.ExecParamsTextOutput{
			Log:  s.Log(),
			Name: "losetup",
			Args: []string{
				"--list",
				"--json",
			},
			StdoutMaxLines: modl.DefaultCommandOutputLines,
		})
	if err != nil {
		return
	}

	var outputModel commandOutput

	if err = json.Unmarshal(bytes.Join(output, nil), &outputModel); err != nil {
		return res, fmt.Errorf("failed to unmarsal command output as JSON; err: %w", err)
	}

	res, err = outputModel.toLoopDeviceList()

	return
}

func (s *impl) attachUsingRegularFileParams(ctx context.Context, newParams *NewParamsFile, attachParams *AttachParamsFile) (res modl.LoopDeviceInfo, err error) {
	if newParams == nil && attachParams == nil {
		return res, errors.New("either 'newParams' or 'attachParams' must be specified")
	} else if newParams != nil && attachParams != nil {
		return res, errors.New("either 'newParams' or 'attachParams' must be not specified (nil)")
	}

	var (
		backingFile     *os.File
		backingFileName string
	)

	switch attachParams {
	case nil:
		if backingFile, err = osutil.OpenFileSize(ctx, newParams.BackingFile, newParams.Size.Amount(), newParams.RecreateFile); err != nil {
			return
		}

		defer func() {
			closeErr := backingFile.Close()

			if err != nil {
				err = errors.Join(err, os.Remove(newParams.BackingFile))
			}

			err = errors.Join(err, closeErr)
		}()

		backingFileName = newParams.BackingFile
	default:
		backingFileName = attachParams.BackingFile
	}

	output, err := osutil.ExecTextOutput(
		ctx,
		osutil.ExecParamsTextOutput{
			Log:  s.Log(),
			Name: "losetup",
			Args: []string{
				"--find",
				"--show",
				backingFileName,
			},
			StdoutMaxLines:    1,
			RequiresSuperUser: true,
		})
	if err != nil {
		return
	} else if len(output) == 0 {
		return res, errors.New("command output is empty")
	}

	name := string(output[0])

	if len(name) == 0 {
		return res, errors.New("no new loop device name returned")
	}

	defer func() {
		if err == nil {
			return
		}

		err = errors.Join(err, s.Detach(ctx, name))
	}()

	list, err := s.List(ctx)
	if err != nil {
		return
	} else if _, j, ok := list.FindByName(name); ok {
		res = list[j]

		return
	}

	return res, fmt.Errorf("created device not found by name '%s' among of results of device listing", name)
}

func New(opts ...OptionFn) (_ Losetup, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", typestring.TypeStringFor[impl](), err)
		}
	}()

	impl := &impl{}
	if err = fopt.Apply(impl, opts...); err != nil {
		return nil, err
	}

	return impl, nil
}
