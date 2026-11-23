package ioext

import (
	"fmt"
	"io/fs"
	"os"
)

type zeroFS struct{}

var zeroFSValue fs.FS = &zeroFS{}

func (*zeroFS) Open(name string) (fs.File, error) {
	return nil, fmt.Errorf("failed to open file '%s' in zero file system (special NOP file system); err: %w", name, os.ErrNotExist)
}

func ZeroFS() fs.FS { return zeroFSValue }
