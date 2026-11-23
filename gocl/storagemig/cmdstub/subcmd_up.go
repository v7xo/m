package cmdstub

import (
	"io/fs"

	"rn.m.vo7.io/vo7io/m/gocl/appcmd/fluentcmd"
)

func setupUp(sourcesDir string, sources fs.FS, cmdOpts fluentcmd.OptionFn[*Conf]) func() (fluentcmd.Cmd, error) {
	return func() (fluentcmd.Cmd, error) {
		return fluentcmd.
			New(fluentcmd.
				Option[*Conf]().
				Use(upSubCmdName).
				Short(upSubcmdShortDesc).
				Conf(
					&Conf{
						Sources:    sources,
						SourcesDir: sourcesDir,
					},
				).
				ExecFn(subcmdExec(SubCmdCodeUp)),
				cmdOpts)
	}
}
