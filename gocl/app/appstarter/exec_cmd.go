package appstarter

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"rn.m.vo7.io/vo7io/m/gocl/appcmd/fluentcmd"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/uuidext"
)

func (s *starter) executeCmd(cobraCmd *cobra.Command, args []string) (err error) {
	cmd, err := s.requireCmd(cobraCmd)
	if err != nil {
		return err
	}

	if cmd == nil {
		return nil
	}

	p := fluentcmd.ExecParams{
		RunTimestamp: s.runTimestamp,
		CalledAs:     s.cmdCalledAs(cobraCmd),
	}

	if p.RunID, err = uuidext.Rand4BHex(); err != nil {
		return fmt.Errorf("failed to generate command run ID; %w", err)
	}

	return cmd.Exec(cobraCmd.Context(), p)
}

func (s *starter) requireCmd(cobraCmd *cobra.Command) (_ fluentcmd.Cmd, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("unexpected condition; no suitable command found to execute; %w", err)
		}
	}()

	var res, nextExpectedCmd fluentcmd.Cmd

	for i := 0; ; i++ {
		if cobraCmd == nil {
			return nil, errors.New("unexpected miss of Cobra command (nil)")
		}

		cmd, ok := s.cmdMap[cobraCmd]
		if ok {
			if i == 0 {
				res = cmd
				nextExpectedCmd = cmd.Parent()
				cobraCmd = cobraCmd.Parent()

				continue
			}

			if nextExpectedCmd != cmd {
				return nil, fmt.Errorf("unexpected match '%s' to '%s' as next parent (Cobra command '%s')", stringsext.Stringer(nextExpectedCmd), stringsext.Stringer(cmd), cobraCmd.CommandPath())
			}

			nextExpectedCmd = cmd.Parent()
			cobraCmd = cobraCmd.Parent()
		} else if cobraCmd == s.rootCobraCmd {
			return res, nil
		} else {
			return nil, errors.New("command map not contains required key")
		}
	}
}
