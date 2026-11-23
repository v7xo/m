package lsblk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/osutil"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/typestring"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type impl struct {
	fopt.LogOpt
}

func (s *impl) Get(ctx context.Context, filt Filter) (res model.BlockDeviceInfo, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to identify block device using filter '%s'; %w", filt, err)

			s.Log().Err(ctx, err).Send()

			return
		}

		s.Log().Trace(ctx).Msgf("succeeded to identify block device using filter '%s'; device: %s", filt, res)
	}()

	list, err := s.List(ctx, filt)
	if err != nil {
		return
	} else if len(list) == 0 {
		return model.BlockDeviceInfo{}, errs.ErrNotFound
	} else if len(list) != 1 {
		return model.BlockDeviceInfo{}, fmt.Errorf("%w; '%d' devices found whereas only one expected", errs.ErrAmbiguousMatch, len(list))
	}

	return list[0], nil
}

func (s *impl) List(ctx context.Context, filt Filter) (res model.BlockDeviceInfoList, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to list block devices using  filter '%s'; %w", filt, err)

			s.Log().Err(ctx, err).Send()

			return
		}

		s.Log().Trace(ctx).Msgf("succeeded to list block devices using  filter '%s', list '%s'", filt, res)
	}()

	output, err := osutil.ExecTextOutput(
		ctx,
		osutil.ExecParamsTextOutput{
			Log:  s.Log(),
			Name: "lsblk",
			Args: []string{
				"-nJpao",
				"TYPE,LABEL,PATH,PARTUUID,UUID,FSTYPE,PARTLABEL",
			},
			StdoutMaxLines: model.DefaultCommandOutputLines,
		})
	if err != nil {
		return nil, err
	}

	var outputModel commandOutput

	if err = json.Unmarshal(bytes.Join(output, nil), &outputModel); err != nil {
		return nil, fmt.Errorf("failed to unmarsal command output as JSON; %w", err)
	}

	if res, err = outputModel.toBlockDeviceList(); err != nil {
		return nil, err
	}

	res = filt.Matched(res)

	return res, nil
}

func New(opts ...OptionFn) (_ Lsblk, err error) {
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
