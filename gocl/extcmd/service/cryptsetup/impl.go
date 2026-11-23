package cryptsetup

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/osutil"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/typestring"
	"rn.m.vo7.io/vo7io/m/gocl/misc/uuidext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type impl struct {
	fopt.LogOpt
}

func (s *impl) Close(ctx context.Context, d model.BlockDeviceID) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to close LUKS device '%s'; %w", d, err)

			s.Log().Err(ctx, err).Send()

			return
		}

		s.Log().Trace(ctx).Msgf("succeeded to close LUKS device '%s'", d)
	}()

	mapperName, err := s.mapperNameFromDeviceID(d)
	if err != nil {
		return
	}

	cmdArgs := []string{"close", mapperName}

	_, err = osutil.ExecTextOutput(
		ctx,
		osutil.ExecParamsTextOutput{
			Log:               s.Log(),
			Name:              cryptsetupCmdName,
			Args:              cmdArgs,
			StdoutMaxLines:    maxCommandOutputBuffer,
			RequiresSuperUser: true,
		})

	return
}

func (s *impl) UUID(ctx context.Context, d model.BlockDeviceID) (res uuid.UUID, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to retrieve UUID of LUKS device '%s' using '%s' command; %w", stringsext.Stringer(d), cryptsetupCmdName, err)

			s.Log().Err(ctx, err).Send()

			return
		}

		s.Log().Trace(ctx).Msgf("succeeded to retrieve UUID of LUKS device '%s'", stringsext.Stringer(d))
	}()

	if d == nil {
		return res, errors.New("device not specified")
	}

	cmdArgs := []string{"luksDump", d.Path()}

	const minOutputLines = 9

	output, err := osutil.ExecTextOutput(
		ctx,
		osutil.ExecParamsTextOutput{
			Log:               s.Log(),
			Name:              cryptsetupCmdName,
			Args:              cmdArgs,
			StdoutMaxLines:    maxCommandOutputBuffer,
			RequiresSuperUser: true,
			StdoutMinLines:    minOutputLines,
		})

	invalidOutputErr := errors.New("invalid command output text format")

	for i, lineBytes := range output {
		line := string(lineBytes)

		//nolint:mnd,revive
		switch i {
		case 0:
			if line != "LUKS header information" {
				return res, invalidOutputErr
			}
		case 1:
			if !strings.HasPrefix(line, "Version:") {
				return res, invalidOutputErr
			}
		case 2:
			if !strings.HasPrefix(line, "Epoch:") {
				return res, invalidOutputErr
			}
		case 3:
			if !strings.HasPrefix(line, "Keyslots area:") {
				return res, invalidOutputErr
			}
		case 4:
			const (
				uuidTextLen = 36
				minLineLen  = uuidTextLen + 15 // UUID: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx

			)
			if !strings.HasPrefix(line, "UUID:") || len(line) < minLineLen {
				return res, invalidOutputErr
			}

			if res, err = uuid.Parse(line[len(line)-uuidTextLen:]); err != nil {
				return
			}
		case 5:
			if !strings.HasPrefix(line, "Label:") {
				return res, invalidOutputErr
			}
		case 6:
			if !strings.HasPrefix(line, "Subsystem:") {
				return res, invalidOutputErr
			}
		case 7:
			if !strings.HasPrefix(line, "Flags:") {
				return res, invalidOutputErr
			}
		}

		//nolint:mnd,revive
		if i > 6 {
			break
		}
	}

	return
}

func (s *impl) Status(ctx context.Context, d model.BlockDeviceID) (status model.CryptDeviceStatus, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to retrieve status of LUKS device '%s' using '%s' command; %w", stringsext.Stringer(d), cryptsetupCmdName, err)

			return
		}

		s.Log().Trace(ctx).Msgf("status of LUKS device '%s' retrieved successfully; status: %s", d, status)
	}()

	mapperName, err := s.mapperNameFromDeviceID(d)
	if err != nil {
		return
	}

	cmdArgs := []string{"status", mapperName}

	output, err := osutil.ExecTextOutput(
		ctx,
		osutil.ExecParamsTextOutput{
			Log:               s.Log(),
			Name:              cryptsetupCmdName,
			Args:              cmdArgs,
			StdoutMaxLines:    maxCommandOutputBuffer,
			RequiresSuperUser: true,
		})

	if err = status.FromCryptsetupStatusOutput(output, d); err != nil {
		return
	}

	return
}

func (s *impl) Open(ctx context.Context, p OpenParams) (d model.BlockDeviceID, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to open LUKS device using params '%s'; %w", stringsext.Stringer(p), err)

			s.Log().Err(ctx, err).Send()

			return
		}

		s.Log().Trace(ctx).Msgf("succeeded to open LUKS device using params '%s'", p)
	}()

	return s.open(ctx, p)
}

func (s *impl) Format(ctx context.Context, params FormatParams) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to format LUKS device using params '%s'; %w", params, err)

			return
		}

		s.Log().Trace(ctx).Msgf("LUKS device formatted successfully using params '%s'", params)
	}()

	switch typed := params.(type) {
	case FormatParamsBase:
		return s.formatUsingBase(ctx, typed)
	case nil:
		return errors.New("nil parameters")
	default:
		return fmt.Errorf("unknown parameters type '%T'", params)
	}
}

func (*impl) mapperNameFromDeviceID(d model.BlockDeviceID) (string, error) {
	if d == nil {
		return "", fmt.Errorf("no device ID specified; %w", errs.ErrNilValue)
	} else if d.Path() == "" {
		return "", fmt.Errorf("specified device ID '%s' have not a path", d)
	} else if d.IsMapper() {
		return d.MapperName(), nil
	}

	devicePath := d.Path()

	if !strings.HasPrefix(devicePath, luksDeviceMapperPathPrefix) {
		return "", fmt.Errorf("invalid device path, missing prefix '%s'", luksDeviceMapperPathPrefix)
	}

	deviceName := filepath.Base(devicePath)

	if deviceName == "" {
		return "", fmt.Errorf("invalid device path '%s', empty device name", devicePath)
	}

	return deviceName, nil
}

func (s *impl) open(ctx context.Context, params OpenParams) (model.BlockDeviceID, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	devPath := params.BackingDevice.Val.Path()
	if devPath == "" {
		return nil, fmt.Errorf("specified backing device ID '%s' not supplies a device path", params.BackingDevice)
	}

	mapperName := params.MapperName
	switch {
	case mapperName == "":
		var mapperNameErr error

		if mapperName, mapperNameErr = uuidext.Rand8BHex(); mapperNameErr != nil {
			return nil, mapperNameErr
		}

		mapperName = "v7xo_" + mapperName
	}

	cmdArgs := make([]string, 0, 3+params.Options.Count()) // nolint:mnd,revive

	cmdArgs = append(cmdArgs, "open")
	cmdArgs = append(cmdArgs, params.Options.Flags("--")...)
	cmdArgs = append(cmdArgs, devPath, mapperName)

	_, err := osutil.ExecTextOutput(
		ctx,
		osutil.ExecParamsTextOutput{
			Log:               s.Log(),
			Name:              cryptsetupCmdName,
			Args:              cmdArgs,
			StdoutMaxLines:    maxCommandOutputBuffer,
			RequiresSuperUser: true,
		})
	if err != nil {
		return nil, err
	}

	dev := model.NewBlockDeviceIDV2(model.BlockDeviceIDOption().MapperName(mapperName).Type(model.BlockDeviceTypeCodeCrypt).Backing(params.BackingDevice.Val))

	return dev, nil
}

func (s *impl) formatUsingBase(ctx context.Context, params FormatParamsBase) (err error) {
	if err = params.Validate(); err != nil {
		return err
	}

	if len(params.BackingDevice.Path()) == 0 {
		return fmt.Errorf("specified backing device ID '%s' have not a path", params.BackingDevice)
	}

	cmdArgs := []string{
		"--batch-mode",
		"--cipher", params.Cipher.Encode(),
		"--disable-external-tokens",
		"--disable-keyring",
		"--hash", "sha512",
		"--key-file", params.KeyFile,
		"--keyslot-cipher", params.Cipher.Encode(),
		"--keyslot-key-size", "512",
		"--label", params.Label,
		"--pbkdf", "pbkdf2",
		"--type", luks2DeviceType,
		"--uuid", params.UUID.String(),
	}

	if params.KeySlot != nil {
		cmdArgs = append(cmdArgs, "--key-slot", stringsext.FormatUintPtr(params.KeySlot))
	}

	if params.PBKDFForceIterations != nil {
		cmdArgs = append(cmdArgs, "--pbkdf-force-iterations", stringsext.FormatUintPtr(params.PBKDFForceIterations))
	}

	if params.PBKDFParallelism != nil {
		cmdArgs = append(cmdArgs, "--pbkdf-parallel", stringsext.FormatUintPtr(params.PBKDFParallelism))
	}

	cmdArgs = append(cmdArgs, "luksFormat", params.BackingDevice.Path())

	s.Log().Info(ctx).Msgf("start of LUKS device formatting; command args: %s", strings.Join(cmdArgs, " "))

	_, err = osutil.ExecTextOutput(
		ctx,
		osutil.ExecParamsTextOutput{
			Log:               s.Log(),
			Name:              cryptsetupCmdName,
			Args:              cmdArgs,
			StdoutMaxLines:    maxCommandOutputBuffer,
			RequiresSuperUser: true,
		})

	return
}

func New(opts ...OptionFn) (_ Cryptsetup, err error) {
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
