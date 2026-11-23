package cmdstub

import (
	"io/fs"

	"rn.m.vo7.io/vo7io/m/gocl/appcmd/fluentcmd"
)

func setupInit(sourcesDir string, sources fs.FS, cmdOpts fluentcmd.OptionFn[*Conf]) func() (fluentcmd.Cmd, error) {
	return func() (fluentcmd.Cmd, error) {
		return fluentcmd.
			New(fluentcmd.
				Option[*Conf]().
				Use(initSubCmdName).
				Short(initSubcmdShortDesc).
				Conf(
					&Conf{
						Sources:    sources,
						SourcesDir: sourcesDir,
					},
				).
				ExecFn(subcmdExec(SubCmdCodeInit)),
				cmdOpts)
	}
}
