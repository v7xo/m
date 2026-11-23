package osutil

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	execpkg "os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/ioext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type ExecParams interface {
	isExecParams()

	log() logext.Logger
	name() string
	args() []string
	requiresSU() bool
	stdin() io.Reader

	Validate() error
	String() string
}

type ExecParamsNullOutput struct {
	Log               logext.Logger
	Stdin             io.Reader
	Name              string
	Args              []string
	RequiresSuperUser bool
}

var _ ExecParams = ExecParamsNullOutput{}

func (v ExecParamsNullOutput) args() []string     { return v.Args }
func (v ExecParamsNullOutput) isExecParams()      {}
func (v ExecParamsNullOutput) log() logext.Logger { return v.Log }
func (v ExecParamsNullOutput) name() string       { return v.Name }
func (v ExecParamsNullOutput) requiresSU() bool   { return v.RequiresSuperUser }
func (v ExecParamsNullOutput) stdin() io.Reader   { return v.Stdin }

func (v ExecParamsNullOutput) Validate() (err error) {
	defer func() {
		if err != nil {
			err = errs.BadArgsWrapWith(fmt.Errorf("specified command execution params not valid; params: %s (%[1]T); %w", v, err))
		}
	}()

	if len(v.Name) == 0 {
		return errors.New("no command name specified")
	}

	return
}

func (v ExecParamsNullOutput) String() string {
	return execParamsString(v.RequiresSuperUser, v.Name, v.Args)
}

type ExecParamsRawOutput struct {
	Log               logext.Logger
	Stdout            io.Writer
	Stdin             io.Reader
	Name              string
	Args              []string
	RequiresSuperUser bool
}

var _ ExecParams = ExecParamsRawOutput{}

func (v ExecParamsRawOutput) args() []string     { return v.Args }
func (v ExecParamsRawOutput) isExecParams()      {}
func (v ExecParamsRawOutput) log() logext.Logger { return v.Log }
func (v ExecParamsRawOutput) name() string       { return v.Name }
func (v ExecParamsRawOutput) requiresSU() bool   { return v.RequiresSuperUser }
func (v ExecParamsRawOutput) stdin() io.Reader   { return v.Stdin }

func (v ExecParamsRawOutput) Validate() (err error) {
	defer func() {
		if err != nil {
			err = errs.BadArgsWrapWith(fmt.Errorf("specified command execution params not valid; params: %s (%[1]T); %w", v, err))
		}
	}()

	if len(v.Name) == 0 {
		return errors.New("no command name specified")
	}

	if v.Stdout == nil {
		return errors.New("no standard output writer specified")
	}

	return
}

func (v ExecParamsRawOutput) String() string {
	return execParamsString(v.RequiresSuperUser, v.Name, v.Args)
}

type ExecParamsTextOutput struct {
	Log   logext.Logger
	Stdin io.Reader
	// Name represents command name.
	Name string
	// Args specifies command arguments.
	Args []string
	// StdoutMinLines specifies min. count of lines (tokens) expected in the command output.
	//
	// Special value for unlimited count not defined.
	StdoutMinLines int
	// StdoutMaxLines specifies max. count of lines (tokens) expected in the command output.
	//
	// Special value for unlimited count not defined.
	StdoutMaxLines int
	// NilTerminator specifies whether to use zero byte as line (token) delimiter in the command output.
	NilTerminator bool
	// RequiresSuperUser indicates whether try to run with super user (for ex., using 'sudo' program).
	RequiresSuperUser bool
}

var _ ExecParams = ExecParamsTextOutput{}

func (v ExecParamsTextOutput) args() []string     { return v.Args }
func (v ExecParamsTextOutput) isExecParams()      {}
func (v ExecParamsTextOutput) log() logext.Logger { return v.Log }
func (v ExecParamsTextOutput) name() string       { return v.Name }
func (v ExecParamsTextOutput) requiresSU() bool   { return v.RequiresSuperUser }
func (v ExecParamsTextOutput) stdin() io.Reader   { return v.Stdin }

func (v ExecParamsTextOutput) Validate() (err error) {
	defer func() {
		if err != nil {
			err = errs.BadArgsWrapWith(fmt.Errorf("specified command execution params not valid; params: %s (%[1]T); %w", v, err))
		}
	}()

	if v.StdoutMinLines > v.StdoutMaxLines {
		err = errors.New("value of min. lines can't be greater than value of max. lines")

		return
	} else if v.StdoutMinLines < 0 {
		err = errors.New("value of min. lines can't be less than zero")

		return
	} else if len(v.Name) == 0 {
		err = errors.New("no command name specified")

		return
	}

	return
}

func (v ExecParamsTextOutput) String() string {
	return execParamsString(v.RequiresSuperUser, v.Name, v.Args)
}

func ExecTextOutput(ctx context.Context, p ExecParamsTextOutput) (res [][]byte, err error) {
	if err = p.Validate(); err != nil {
		return
	}

	if p.StdoutMaxLines > 0 {
		res = make([][]byte, 0, p.StdoutMaxLines)
	}

	err = exec(
		ctx,
		p,
		func(b []byte) error {
			if b == nil {
				res = append(res, nil)
			} else if len(b) == 0 {
				res = append(res, []byte{})
			} else {
				cp := make([]byte, len(b))

				copy(cp, b)

				res = append(res, cp)
			}

			return nil
		},
	)

	return
}

func ExecTextOutputReader(ctx context.Context, p ExecParamsTextOutput) (res io.Reader, err error) {
	if err = p.Validate(); err != nil {
		return
	}

	buf := bytes.NewBuffer(make([]byte, 0, ioext.DefaultApplicationBufferSize))

	return buf,
		exec(
			ctx,
			p,
			func(line []byte) error {
				switch {
				case len(line) != 0:
					w, writeErr := buf.Write(line)
					if writeErr != nil {
						return fmt.Errorf("failed to write output to bytes buffer; %w", writeErr)
					} else if w != len(line) {
						return fmt.Errorf("incompleted write of output to bytes buffer, count of written bytes '%d' not equals to '%d'", w, len(line))
					}
				}

				if err = buf.WriteByte('\n'); err != nil {
					return fmt.Errorf("failed to write output to bytes buffer; %w", err)
				}

				return nil
			},
		)
}

func Exec(ctx context.Context, p ExecParams) (err error) {
	if err = p.Validate(); err != nil {
		return
	}

	return exec(
		ctx,
		p,
		func(_ []byte) error {
			return nil // TODO: write to log optionally.
		},
	)
}

func exec(ctx context.Context, params ExecParams, tokenHandler func([]byte) error) (err error) {
	log := params.log()
	if log == nil {
		if log, err = contextext.Log(ctx); err != nil {
			return err
		}

		if log == nil {
			log = logext.NopLogger
		}
	}

	defer func() {
		if err != nil || log.IsNop() {
			return
		}

		log.Trace(ctx).Str(attrname.OSExecCmd, params.String()).Msgf("succeeded to execute command '%s'", params.name())
	}()

	// TODO: replace with limited buffer.
	stderrBuf := bytes.NewBuffer(make([]byte, 0, 2048)) // nolint:mnd,revive

	defer func() {
		stderrText := strings.TrimRight(stderrBuf.String(), "\n")

		if err != nil {
			err = fmt.Errorf("failed to execute command '%s'; %w; Stderr output: %s", params.name(), err, placeholder.UnspecifiedIfEmpty(stderrText))

			log.Err(ctx, err).Str(attrname.OSExecCmd, params.String()).Send()

			return
		}

		if stderrText != "" && !log.IsNop() {
			log.Warn(ctx).Str(attrname.OSExecCmd, params.String()).Msg(stderrText)
		}
	}()

	name := params.name()
	args := make([]string, 0, 1+len(params.args())) // nolint:mnd,revive

	switch {
	case params.requiresSU() && os.Geteuid() != 0:
		var (
			lookErr     error
			cmdFullPath string
		)

		if _, lookErr = execpkg.LookPath("sudo"); lookErr == nil {
			name = "sudo"
			args = append(args, params.name())
			args = append(args, params.args()...)
		} else {
			log.Warn(ctx).Err(fmt.Errorf("failed to look path for 'sudo'; %w", lookErr)).Send()

			cmdFullPath, lookErr = execpkg.LookPath(name)
			if lookErr != nil {
				return fmt.Errorf("failed to look path for '%s'; %w", name, lookErr)
			}

			if cmdFullPath, lookErr = filepath.EvalSymlinks(cmdFullPath); lookErr != nil {
				return fmt.Errorf("failed to evaluate sym links for '%s'; %w", cmdFullPath, lookErr)
			}

			name = "pkexec"

			args = append(args, "--keep-cwd")
			args = append(args, cmdFullPath)
			args = append(args, params.args()...)
		}
	default:
		args = params.args()
	}

	cmd := execpkg.CommandContext(ctx, name, args...) // nolint:gosec
	cmd.Stderr = stderrBuf

	if params.stdin() != nil {
		cmd.Stdin = params.stdin()
	}

	switch paramsTyped := params.(type) {
	case ExecParamsNullOutput:
		cmd.Stdout = io.Discard
		err = cmd.Run()
	case ExecParamsRawOutput:
		cmd.Stdout = paramsTyped.Stdout
		err = cmd.Run()
	case ExecParamsTextOutput:
		stdoutBuf := bytes.NewBuffer(nil) // TODO: replace with limited buffer.
		cmd.Stdout = stdoutBuf

		if err = cmd.Run(); err != nil {
			return
		}

		s := bufio.NewScanner(stdoutBuf)

		if paramsTyped.NilTerminator {
			s.Split(scanZeroByteLines)
		}

		cnt := 0

		for s.Scan() {
			cnt++

			if cnt > paramsTyped.StdoutMaxLines {
				return fmt.Errorf("next token read from command standard output whereas it not expected; expected max. token count: %d", paramsTyped.StdoutMaxLines)
			}

			t := s.Bytes()

			if handlerErr := tokenHandler(t); handlerErr != nil {
				return fmt.Errorf("failed to handle next token from command standard output; token length: %d; %w", len(t), handlerErr)
			}
		}

		if s.Err() != nil {
			return fmt.Errorf("error occurred while scanning command standard output; %w", err)
		} else if cnt < paramsTyped.StdoutMinLines {
			return fmt.Errorf("command standard output contains '%d' tokens that is less than expected '%d'", cnt, paramsTyped.StdoutMinLines)
		}
	default:
		return fmt.Errorf("invalid type '%T' of execution params", params)
	}

	return
}

func scanZeroByteLines(d []byte, eof bool) (advance int, token []byte, err error) {
	if eof && len(d) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(d, 0); i >= 0 {
		return i + 1, d[0:i], nil
	} else if eof {
		return len(d), d, nil
	}

	return 0, nil, nil
}

func execParamsString(reqSU bool, n string, args []string) string {
	sb := strings.Builder{}
	q := func(val string) string {
		buf := make([]rune, 0, 32) // nolint:mnd,revive
		needQuote := false

		for _, sym := range val {
			switch {
			case sym == '"' || sym == '\\':
				buf = append(buf, '\\', sym)

				needQuote = true
			case unicode.IsSpace(sym):
				buf = append(buf, sym)

				needQuote = true
			default:
				buf = append(buf, sym)
			}
		}

		if needQuote {
			buf = append(buf, '"', '"')

			copy(buf[1:], buf)

			buf[0] = '"'
		}

		return string(buf)
	}

	sb.WriteByte('(')

	sb.WriteString("requires su: ")
	sb.WriteString(strconv.FormatBool(reqSU))

	sb.WriteString(", [")

	sb.WriteString(q(n))

	for i := range args {
		sb.WriteByte(' ')
		sb.WriteString(q(args[i]))
	}

	sb.WriteString("])")

	return sb.String()
}
