package fluentcmd

import (
	"context"
)

// Cmd represents command interface.
type Cmd interface {
	Children() []Cmd
	Parent() Cmd
	SetParent(Cmd) error
	Short() string
	SilenceError() bool
	SilenceUsage() bool
	String() string
	Use() string

	Exec(context.Context, ExecParams) error
}

type CmdGenerator interface{ GenerateCmd() (Cmd, error) }
