package stdcase

import (
	"io/fs"
)

type MigratedStorageProps struct {
	Sources1 fs.FS
	Sources2 fs.FS
	Sources3 fs.FS
	Sources4 fs.FS

	DeployJobID string
}
