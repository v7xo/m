package appstarter

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"rn.m.vo7.io/vo7io/m/gocl/appcmd/fluentcmd"
	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

type OptionFn func(*starter) error

func (o OptionFn) CmdFn(val func() (fluentcmd.Cmd, error)) OptionFn {
	return func(s *starter) error {
		if val == nil {
			return errors.New("command generator func can't be a nil")
		}

		if err := o(s); err != nil {
			return err
		}

		cmd, err := val()
		if err != nil {
			return fmt.Errorf("command generator func has returned an error as a result; %w", err)
		}

		if cmd == nil {
			return errors.New("command generator func has returned a nil as a result")
		}

		type traversalE struct {
			cmd         fluentcmd.Cmd
			cobraParent *cobra.Command
		}

		traversal := []traversalE{
			{
				cmd:         cmd,
				cobraParent: s.rootCobraCmd,
			},
		}

		for len(traversal) != 0 {
			e := traversal[0]
			traversal = traversal[1:]

			flset, err := o.generateFlagSet(e.cmd)
			if err != nil {
				return err
			}

			cmdCobraDelegate := &cobra.Command{
				Args:          cobra.NoArgs,
				RunE:          s.executeCmd,
				Short:         e.cmd.Short(),
				SilenceErrors: e.cmd.SilenceError(),
				SilenceUsage:  e.cmd.SilenceUsage(),
				Use:           e.cmd.Use(),
			}

			cmdCobraDelegate.Flags().AddFlagSet(flset)

			s.cmdMap[cmdCobraDelegate] = e.cmd
			e.cobraParent.AddCommand(cmdCobraDelegate)

			for _, child := range e.cmd.Children() {
				traversal = append(traversal, traversalE{cmd: child, cobraParent: cmdCobraDelegate})
			}
		}

		return nil
	}
}

func (o OptionFn) Cmd(val fluentcmd.Cmd) OptionFn {
	return o.CmdFn(func() (fluentcmd.Cmd, error) { return val, nil })
}

//nolint:revive
func (OptionFn) generateFlagSet(cmd fluentcmd.Cmd) (*pflag.FlagSet, error) {
	if gen, ok := cmd.(flenv.FlagSetGenerator); !ok {
		return nil, nil
	} else if flset, err := gen.GenerateFlagSet(); err != nil {
		return nil, err
	} else {
		return flset, nil
	}
}

func Option() OptionFn { return fopt.Nop[starter] }
