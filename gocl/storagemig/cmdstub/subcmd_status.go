package cmdstub

import (
	"io/fs"

	"rn.m.vo7.io/vo7io/m/gocl/appcmd/fluentcmd"
)

func setupStatus(sourcesDir string, sources fs.FS, cmdOpts fluentcmd.OptionFn[*Conf]) func() (fluentcmd.Cmd, error) {
	return func() (fluentcmd.Cmd, error) {
		return fluentcmd.
			New(fluentcmd.
				Option[*Conf]().
				Conf(
					&Conf{
						Sources:    sources,
						SourcesDir: sourcesDir,
					},
				).
				Use(statusSubCmdName).
				Short(statusSubcmdShortDesc).
				ExecFn(subcmdExec(SubCmdCodeStatus)),
				cmdOpts)
	}
}
