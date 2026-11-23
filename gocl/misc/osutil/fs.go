package osutil

import (
	"context"
	crypt_rand "crypto/rand"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"rn.m.vo7.io/vo7io/m/gocl/misc/contextext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/ioext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/uuidext"
)

func Exists(n string) (bool, error) {
	_, err := os.Stat(n)

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}

	return err == nil, err
}

func DirExists(n string) (bool, error) {
	stat, err := os.Stat(n)

	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	} else if stat.IsDir() {
		return true, nil
	}

	return false, fmt.Errorf("path '%s' exists and not points to a directory as it expected", n)
}

func FileExists(n string) (exists bool, stat os.FileInfo, err error) {
	stat, err = os.Stat(n)

	if errors.Is(err, os.ErrNotExist) {
		return false, nil, nil
	}

	if err != nil {
		return false, nil, err
	}

	if stat.Mode().IsRegular() || (stat.Mode()&os.ModeSymlink) == os.ModeSymlink {
		return true, stat, nil
	}

	return false, nil, fmt.Errorf("path '%s' exists and not points to a file as it expected", n)
}

func OpenFileSize(ctx context.Context, n string, requiredSize int64, rewrite bool) (f *os.File, err error) {
	flags := os.O_RDWR | os.O_CREATE
	if rewrite {
		flags |= os.O_TRUNC
	}

	f, err = os.OpenFile(n, flags, 0o666) //nolint:revive,mnd
	if err != nil {
		return nil, fmt.Errorf("error while opening file '%s'; err: %w", n, err)
	}

	defer func() {
		if err != nil {
			if closeErr := f.Close(); closeErr != nil {
				err = fmt.Errorf("%w; %w", err, closeErr)
			}
		}
	}()

	switch {
	case !rewrite:
		stat, statErr := f.Stat()
		if statErr != nil {
			return nil, fmt.Errorf("error while reading info about file '%s'; err: %w", n, statErr)
		} else if stat.Size() == requiredSize {
			return f, nil
		} else if stat.Size() != 0 {
			return nil,
				fmt.Errorf("file '%s' already exists and file rewriting prohibited by params,"+
					" but size '%d' of existing file not equals to required '%d' as it specified by params ", n, stat.Size(), requiredSize)
		}
	}

	log, _ := contextext.Log(ctx) //nolint
	if log != nil && !log.IsNop() {
		requiredSize := requiredSize

		defer func() {
			if err != nil {
				return
			}

			log.Info(ctx).Msgf("file '%s' with required '%d' size created/opened, its content overwritten with random bytes", n, requiredSize)
		}()
	}

	bs := min(ioext.DefaultApplicationBufferSize, requiredSize)
	b := make([]byte, bs)

	for requiredSize > 0 {
		select {
		case <-ctx.Done():
			return f, ctx.Err()
		default:
		}

		bs = min(bs, requiredSize)
		b = b[:bs]

		if readCount, readErr := crypt_rand.Read(b); readErr != nil {
			return f, readErr
		} else if readCount != len(b) {
			return f, fmt.Errorf("incompleted buffer fill with random bytes, expected '%d', actual '%d'", len(b), readCount)
		}

		if _, err = f.Write(b); err != nil {
			return f, err
		}

		requiredSize -= int64(len(b))
	}

	return
}

func AbsTempDir() (string, error) {
	tmpDir := os.TempDir()
	if tmpDir == "" {
		return "", errors.New("empty path of temporary directory returned as a result of os.TempDir()")
	}

	if filepath.IsAbs(tmpDir) {
		return tmpDir, nil
	}

	absTmpDir, absErr := filepath.Abs(tmpDir)
	if absErr != nil {
		return "", fmt.Errorf("failed to make path '%s' absolute; %w", tmpDir, absErr)
	}

	return absTmpDir, nil
}

func MkdirAllRand(parentDir string, prefix string, perm os.FileMode) (string, error) {
	randPart, err := uuidext.Rand4BHex()
	if err != nil {
		return "", err
	}

	n := prefix
	if n == "" {
		n = randPart
	} else {
		n = n + "_" + randPart
	}

	path := filepath.Join(parentDir, n)
	if err = os.MkdirAll(path, perm); err != nil {
		return "", err
	}

	return path, nil
}

func MkdirTemp(pattern string) (_ string, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to make a temporary directory using pattern '%s'; %w", pattern, err)
		}
	}()

	tmpDir, err := AbsTempDir()
	if err != nil {
		return "", err
	}

	tmpDirMakeTried := false

	for {
		var res string

		if res, err = os.MkdirTemp(tmpDir, pattern); err != nil {
			switch {
			case os.IsNotExist(err) && !tmpDirMakeTried:
				tmpDirMakeTried = true

				if err = os.MkdirAll(tmpDir, 0o700); err != nil {
					return "", err
				}

				continue
			default:
				return "", err
			}
		}

		return res, nil
	}
}
