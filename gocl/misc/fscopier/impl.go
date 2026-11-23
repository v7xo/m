package fscopier

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"golang.org/x/sync/errgroup"

	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type impl struct {
	fopt.LogOpt

	pathFilter     PathFilterFn
	existingFilter ExistingFilterFn
	filePermFilter PermFilter
	contentFilter  ContentFilterFn
	nameFilter     NameFilterFn
	existingAction ExistingAction

	disableDeletedSuffix  bool
	disableNameUnescaping bool
}

//nolint:maintidx
func (s *impl) CopyFileSystem(ctx context.Context, src fs.FS, srcDir, dstDir string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("error while copying file system (type: %s) from '%s' to '%s'; %w", stringsext.TypeStringOfVal(src), srcDir, dstDir, err)
		}
	}()

	srcDirInfo, err := fs.Stat(src, srcDir)
	if err != nil {
		return err
	} else if !srcDirInfo.IsDir() {
		return errors.New("src is not a directory")
	}

	type traversalItem struct {
		srcPath string
		dstPath string
	}

	srcTraversal := []traversalItem{
		{
			srcPath: srcDir,
			dstPath: dstDir,
		},
	}

	copyItems := []copyItem{}

	for len(srcTraversal) != 0 {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		default:
		}

		srcEntry := srcTraversal[0]

		copyItems = append(
			copyItems,
			copyItem{
				dstPath: srcEntry.dstPath,
				isDir:   true,
				srcPath: srcEntry.srcPath,
			},
		)

		var srcChildren []fs.DirEntry

		srcChildren, err = fs.ReadDir(src, srcEntry.srcPath)
		if err != nil {
			return fmt.Errorf("error while reading dirs; path: %s; %w", srcEntry.srcPath, err)
		}

		for _, srcChild := range srcChildren {
			select {
			case <-ctx.Done():
				return context.Cause(ctx)
			default:
			}

			srcChildPath := filepath.Join(srcEntry.srcPath, srcChild.Name())

			if srcChild.IsDir() {
				mappedDstDirPath, deleteDstDir, mapDstErr := s.mapDstPath(srcChildPath, srcChild.Name(), srcEntry.dstPath)
				if mapDstErr != nil {
					return mapDstErr
				}

				s.Log().Trace(ctx).Msgf("map name: %s -> %s", srcChildPath, mappedDstDirPath)

				if deleteDstDir {
					copyItems = append(
						copyItems,
						copyItem{
							srcPath:      srcChildPath,
							isDir:        true,
							dstPath:      mappedDstDirPath,
							dstMustBeDel: true,
						},
					)

					continue
				}

				var nonIgnored bool

				if nonIgnored, err = s.applyPathFilter(srcChildPath, mappedDstDirPath, false); err != nil {
					return err
				} else if !nonIgnored {
					copyItems = append(
						copyItems,
						copyItem{
							srcPath:       srcChildPath,
							isDir:         true,
							dstPath:       mappedDstDirPath,
							dstMustBeSkip: true,
						},
					)

					continue
				}

				srcTraversal = append(
					srcTraversal,
					traversalItem{
						srcPath: srcChildPath,
						dstPath: mappedDstDirPath,
					},
				)

				continue
			} else if !srcChild.Type().IsRegular() {
				return fmt.Errorf("src is not a regular file; src: %s", srcChildPath)
			}

			mappedDstFilePath, deleteDstFile, mapDstErr := s.mapDstPath(srcChildPath, srcChild.Name(), srcEntry.dstPath)
			if mapDstErr != nil {
				return mapDstErr
			}

			s.Log().Info(ctx).Msgf("map name: %s -> %s", srcChildPath, mappedDstFilePath)

			if deleteDstFile {
				copyItems = append(
					copyItems,
					copyItem{
						srcPath:      srcChildPath,
						dstPath:      mappedDstFilePath,
						dstMustBeDel: true,
					},
				)

				continue
			}

			var nonIgnored bool

			if nonIgnored, err = s.applyPathFilter(srcChildPath, mappedDstFilePath, true); err != nil {
				return err
			} else if !nonIgnored {
				copyItems = append(
					copyItems,
					copyItem{
						srcPath:       srcChildPath,
						dstPath:       mappedDstFilePath,
						dstMustBeSkip: true,
					},
				)

				continue
			}

			copyItems = append(
				copyItems,
				copyItem{
					srcPath: srcChildPath,
					dstPath: mappedDstFilePath,
				},
			)
		}

		srcTraversal = srcTraversal[1:]
	}

	return s.doCopy(ctx, copyItems, src, dstDir)
}

func (s *impl) doCopy(ctx context.Context, items []copyItem, src fs.FS, dstBase string) error {
	slices.SortFunc(
		items,
		func(a copyItem, b copyItem) int {
			if a.isDir {
				switch {
				case b.isDir:
					return strings.Compare(a.dstPath, b.dstPath)
				default:
					return -1
				}
			} else if b.isDir {
				return 1
			}

			return strings.Compare(a.dstPath, b.dstPath)
		},
	)

	mkdirAll := true

	for len(items) != 0 {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		default:
		}

		if !items[0].isDir {
			break
		}

		switch {
		case items[0].dstMustBeDel:
			if err := os.RemoveAll(items[0].dstPath); err != nil && !errors.Is(err, os.ErrNotExist) {
				return err
			}
		case items[0].dstMustBeSkip:
		default:
			if mkdirAll {
				switch err := os.MkdirAll(items[0].dstPath, dstDirDefaultPerm); err != nil {
				case true:
					return err
				}

				mkdirAll = false
			} else if err := os.Mkdir(items[0].dstPath, dstDirDefaultPerm); err != nil && !errors.Is(err, os.ErrExist) {
				return err
			}
		}

		items = items[1:]
	}

	if len(items) == 0 {
		return nil
	}

	group, groupCtx := errgroup.WithContext(ctx)

	for _, item := range items {
		group.Go(
			func() error {
				switch {
				case item.dstMustBeDel:
					if err := os.Remove(item.dstPath); err != nil && !errors.Is(err, os.ErrNotExist) {
						return err
					}

					return nil
				case item.dstMustBeSkip:
					return nil
				}

				return s.doCopyFile(groupCtx, src, item.srcPath, dstBase, item.dstPath)
			},
		)
	}

	return group.Wait()
}

//nolint:funlen,forbidigo,gocyclo,cyclop,maintidx
func (s *impl) doCopyFile(ctx context.Context, src fs.FS, srcPath, dstBase, dstPath string) (err error) {
	dstFileToPrint, relErr := filepath.Rel(dstBase, dstPath)
	if relErr != nil {
		dstFileToPrint = dstPath
	}

	s.Log().Trace(ctx).Msgf("copy '%s' -> '%s' ...", srcPath, dstFileToPrint)

	var (
		dstRWChmod          bool
		dstFilePerm         fs.FileMode
		dstSkipExistingFile bool
	)

	defer func() {
		err = errs.Join(err, errs.PanicErr(recover()))
		if err == nil {
			if err = os.Chmod(dstPath, dstFilePerm); err != nil {
				err = fmt.Errorf("failed to set permissions '%s' on file '%s'; %w", dstFilePerm, dstPath, err)

				return
			}
		}

		if err != nil {
			s.Log().Err(ctx, err).Msgf("copy '%s' -> '%s' - error", srcPath, dstFileToPrint)

			return
		}

		if dstSkipExistingFile {
			s.Log().Info(ctx).Msgf("copy '%s' -> '%s' - skipped", srcPath, dstFileToPrint)

			return
		}

		s.Log().Info(ctx).Msgf("copy '%s' -> '%s' - ok, overwritten", srcPath, dstFileToPrint)
	}()

	switch {
	case s.filePermFilter != nil:
		dstFilePerm, err = s.filePermFilter.Calc(srcPath, dstPath)
		if err != nil {
			return fmt.Errorf("failed to calculate permissions for destination file '%s'; %w", dstPath, err)
		}
	}

	_, err = os.Stat(dstPath)

	switch {
	case err == nil:
		dstFileAction := s.existingAction

	existingActionChoose:
		switch dstFileAction {
		case ExistingActionOverwrite:
			dstRWChmod = true
		case ExistingActionSkip:
			dstSkipExistingFile = true

			return nil
		case ExistingActionFilter:
			dstFileAction, err = s.existingFilter(srcPath, dstPath)
			if err != nil {
				return fmt.Errorf("error while evaluating action filter for existing file; file: %s; %w", dstPath, err)
			}

			if dstFileAction == ExistingActionFilter {
				return fmt.Errorf("error while evaluating action for existing file; evaluation result '%s' is invalid; file: %s", ExistingActionFilter, dstPath)
			}

			goto existingActionChoose
		case ExistingActionFail:
			return fmt.Errorf("file already exists; file: %s", dstPath)
		default:
			return fmt.Errorf("specified action for existing file not supported; action: %s; file: %s", dstFileAction, dstPath)
		}
	case errors.Is(err, os.ErrNotExist):
		err = nil
	default:
		return err
	}

	// write out file
	//
	srcFileBuf, err := fs.ReadFile(src, srcPath)
	if err != nil {
		return fmt.Errorf("error while reading file; file: %s; %w", srcPath, err)
	}

	var dstFileBuf []byte

	if s.contentFilter == nil {
		dstFileBuf = srcFileBuf
	} else if dstFileBuf, err = s.contentFilter(srcPath, dstPath, srcFileBuf); err != nil {
		return fmt.Errorf("failed to apply file content filter; file: %s; %w", dstPath, err)
	}

	if dstRWChmod {
		if err = os.Chmod(dstPath, dstFileRWPerm); err != nil {
			return fmt.Errorf("error while setting write permission flag on file; file: %s; %w", dstPath, err)
		}
	}

	if err = os.WriteFile(dstPath, dstFileBuf, dstFileRWPerm); err != nil {
		return fmt.Errorf("error while writing out file; file: %s; %w", dstPath, err)
	}

	return nil
}

func (s *impl) applyPathFilter(srcPath, dstDir string, file bool) (bool, error) {
	if s.pathFilter == nil {
		return true, nil
	}

	pathFilterAction, err := s.pathFilter(srcPath, dstDir, file)
	if err != nil {
		return false, err
	}

	switch pathFilterAction {
	case PathFilterActionIgnore:
		return false, nil
	case PathFilterActionDeleteExisting:
		if err := os.RemoveAll(dstDir); err != nil && !errors.Is(err, os.ErrNotExist) {
			return false, err
		}

		return false, nil
	case PathFilterActionContinue:
		return true, nil
	case PathFilterActionUnspecified:
		return false, errors.New("path filter action not specified")
	default:
		return false, fmt.Errorf(
			"path filter evaluated to unknown action; action: %s; src path: %s; dst path: %s",
			pathFilterAction,
			srcPath,
			dstDir,
		)
	}
}

func (s *impl) mapDstPath(srcPath, srcName, dstDir string) (res string, deleted bool, err error) {
	switch {
	case !s.disableDeletedSuffix:
		if srcNameExt := path.Ext(srcName); srcNameExt == DeletedSuffix {
			deleted = true
			srcName = srcName[:len(srcName)-len(srcNameExt)]
		}
	}

	dstName := srcName

	switch {
	case !s.disableNameUnescaping:
		if strings.HasSuffix(dstName, "_") && len(dstName) > 1 {
			dstName = dstName[:len(dstName)-1]
		}

		dstName = strings.ReplaceAll(dstName, DotSymSpecifier, ".")
		dstName = strings.ReplaceAll(dstName, UnderscoreSymSpecifier, "_")
	}

	switch {
	case s.nameFilter != nil:
		if dstName, err = s.nameFilter(srcPath, srcName, dstDir, dstName); err != nil {
			return "", false, fmt.Errorf("error while evaluating name filter; srcDir: %s; srcName: %s; dstDir: %s; dstName: %s", srcPath, srcName, dstDir, dstName)
		}
	}

	return filepath.Join(dstDir, dstName), deleted, nil
}

func New(opts ...OptionFn) (_ FileSystemCopier, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", stringsext.TypeStringFor[impl](), err)
		}
	}()

	impl := &impl{}
	if err = fopt.Apply(impl, opts...); err != nil {
		return nil, err
	}

	if impl.existingAction == ExistingActionFilter && impl.existingFilter == nil {
		return nil, errors.New("existing file filter function not specified")
	}

	return impl, nil
}
