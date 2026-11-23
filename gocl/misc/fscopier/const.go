package fscopier

import "io/fs"

//nolint:gofumpt
const (
	dstDirDefaultPerm  fs.FileMode = 0o744
	dstFileDefaultPerm fs.FileMode = 0o644
	dstFileRWPerm      fs.FileMode = 0o644

	// DeletedSuffix is special name suffix denoting dir/file to be deleted in destination.
	DeletedSuffix = ".deleted"

	LogScope = "cl_fs-copier"
)

// Following consts designate specifiers used in names.
// one of use case - to mask some symbols in the file name (ex., names starting with '.' or '_' ignored by Go embedding).
const (
	// DotSymSpecifier special sequence used to designate '.' symbol in real name.
	DotSymSpecifier = `%dot`
	// UnderscoreSymSpecifier special sequence used to designate '_' symbol in real name.
	UnderscoreSymSpecifier = `%underscore`
)
