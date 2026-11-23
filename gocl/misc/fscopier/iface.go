package fscopier

import (
	"context"
	"io/fs"
)

type FileSystemCopier interface {
	CopyFileSystem(ctx context.Context, src fs.FS, srcDir, dstDir string) error
}
