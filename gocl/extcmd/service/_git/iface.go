package git

import "context"

type Service interface {
	// CheckIgnore checks whether specified path name n is Git-ignored.
	CheckIgnore(ctx context.Context, n string) (bool, error)
	// HeadHash returns hash of commit the HEAD is pointing to.
	HeadHash(ctx context.Context) (string, error)
	// Status returns info about working tree state.
	Status(ctx context.Context) ([]WorkingTreeStatusEntry, error)
	// TopLevel returns top-level directory of working tree directory (if current working directory points to Git working tree directory).
	TopLevelDir(ctx context.Context) (string, error)
}
