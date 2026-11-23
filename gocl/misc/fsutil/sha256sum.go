package fsutil

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	fspkg "io/fs"
	"os"
	"path/filepath"
	"slices"

	"golang.org/x/sync/errgroup"

	"rn.m.vo7.io/vo7io/m/gocl/misc/ioext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/syncext"
)

func DirSHA256Sum(ctx context.Context, fs fspkg.FS, dirName string, dop syncext.MaxDOP, skip func(n string, e fspkg.DirEntry) (bool, error)) (_ []byte, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to compute SHA256 sum of directory '%s' (fs type '%s'); %w", dirName, stringsext.TypeStringOfVal(fs), err)
		}
	}()

	effectiveDOP, err := dop.Effective()
	if err != nil {
		return
	}

	entryDir := dirName

	switch fs {
	case nil:
		var rootFS *os.Root

		if rootFS, err = os.OpenRoot(entryDir); err != nil {
			return nil, fmt.Errorf("failed to open specified directory; %w", err)
		}

		fs = rootFS.FS()
		entryDir = "."
	}

	entryDirInfo, err := fspkg.Stat(fs, entryDir)
	if err != nil {
		return
	} else if !entryDirInfo.IsDir() {
		return nil, fmt.Errorf("specified name '%s' doesn't point to a directory", dirName)
	}

	if skip == nil {
		skip = func(n string, e fspkg.DirEntry) (bool, error) { return false, nil }
	}

	type traversalE struct {
		path       string
		pathToHash string
	}

	const (
		dirMarker     = "D:"
		regFileMarker = "R:"
		rootDirDot    = "."
	)

	traversal := []traversalE{{path: entryDir, pathToHash: ""}}
	pathsToHash := []string{dirMarker + rootDirDot}
	filesToHash := []string{}

	for len(traversal) != 0 {
		select {
		case <-ctx.Done():
			return nil, context.Cause(ctx)
		default:
		}

		entry := traversal[0]

		var children []fspkg.DirEntry

		children, err = fspkg.ReadDir(fs, entry.path)
		if err != nil {
			return nil, fmt.Errorf("error while reading dirs; path: %s; %w", entry.path, err)
		}

		for _, child := range children {
			select {
			case <-ctx.Done():
				return nil, context.Cause(ctx)
			default:
			}

			childPath := filepath.Join(entry.path, child.Name())
			childPathToHash := filepath.Join(entry.pathToHash, child.Name())
			shouldSkip := false

			if shouldSkip, err = skip(childPath, child); err != nil {
				return nil, fmt.Errorf("failed to evaluate skip filter; file: %s; %w", childPath, err)
			} else if shouldSkip {
				continue
			}

			if child.IsDir() {
				traversal = append(traversal, traversalE{path: childPath, pathToHash: childPathToHash})
				pathsToHash = append(pathsToHash, dirMarker+childPathToHash)

				continue
			} else if !child.Type().IsRegular() {
				return nil, fmt.Errorf("not a regular file; file: %s", childPath)
			}

			var childFI fspkg.FileInfo

			if childFI, err = child.Info(); err != nil {
				return nil, fmt.Errorf("failed to get file info; file: %s; %w", childPath, err)
			}

			pathsToHash = append(pathsToHash, regFileMarker+childPathToHash)

			if childFI.Size() != 0 {
				filesToHash = append(filesToHash, childPath)
			}
		}

		traversal = traversal[1:]
	}

	group, _ := errgroup.WithContext(ctx)
	filesHashes := make([][]byte, len(filesToHash))

	group.SetLimit(effectiveDOP)

	if len(filesToHash) != 0 {
		slices.Sort(filesToHash)

		for i := range filesToHash {
			fileToHashIdx := i

			group.Go(
				func() error {
					r, localErr := fs.Open(filesToHash[fileToHashIdx])
					if localErr != nil {
						return fmt.Errorf("error while opening file; file: %s; %w", filesToHash[fileToHashIdx], localErr)
					}

					defer r.Close()

					fileHash, localErr := ioext.SHA256Sum(r)
					if localErr != nil {
						return fmt.Errorf("error while computing hash of file; file: %s; %w", filesToHash[fileToHashIdx], localErr)
					}

					filesHashes[fileToHashIdx] = fileHash

					return nil
				},
			)
		}

		if err = group.Wait(); err != nil {
			return
		}
	}

	slices.Sort(pathsToHash)

	h := sha256.New()

	for _, v := range pathsToHash {
		h.Write([]byte(v))
	}

	for _, v := range filesHashes {
		h.Write(v)
	}

	return h.Sum(nil), nil
}

func DirSHA256SumHex(ctx context.Context, fs fspkg.FS, dir string, maxDOP syncext.MaxDOP, skip func(n string, e fspkg.DirEntry) (bool, error)) (string, error) {
	h, err := DirSHA256Sum(ctx, fs, dir, maxDOP, skip)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h), nil
}

func FileSHA256Sum(n string) ([]byte, error) {
	f, err := os.Open(n)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	return ioext.SHA256Sum(f)
}

func FileSHA256SumHex(n string) (string, error) {
	h, err := FileSHA256Sum(n)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h), nil
}
