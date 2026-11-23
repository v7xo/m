package fluentcmd

import "time"

type ExecParams struct {
	RunTimestamp time.Time
	RunID        string
	CalledAs     string
}
