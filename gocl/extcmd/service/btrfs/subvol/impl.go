package subvol

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"

	extcmd_model "rn.m.vo7.io/vo7io/m/gocl/extcmd/model"
	"rn.m.vo7.io/vo7io/m/gocl/extcmd/service/findmnt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/osutil"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/uuidext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type impl struct {
	fopt.LogOpt

	findmnt findmnt.Findmnt
}

func (s *impl) Transmit(ctx context.Context, p TransmitParams) (Summary, error) {
	return s.transmit(ctx, p)
}

func (s *impl) Send(ctx context.Context, parent, subvol string, w io.Writer) (err error) {
	diffMsg := ""

	if parent != "" {
		diffMsg = fmt.Sprintf(" (differential, relatively '%s')", parent)
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to send BTRFS subvolume(s) '%s'%s; %w", subvol, diffMsg, err)

			s.Log().Err(ctx, err).Send()

			return
		}

		s.Log().Trace(ctx).Msgf("succeeded to send BTRFS subvolume(s) '%s'%s", subvol, diffMsg)
	}()

	cmdArgs := []string{"send", "-q", "-e", "--proto=2"}

	if parent != "" {
		cmdArgs = append(cmdArgs, "-p", parent)
	}

	cmdArgs = append(cmdArgs, subvol)

	return osutil.Exec(
		ctx,
		osutil.ExecParamsRawOutput{
			Args:              cmdArgs,
			Log:               s.Log(),
			Name:              btrfsCmdName,
			RequiresSuperUser: true,
			Stdout:            w,
		},
	)
}

func (s *impl) Receive(ctx context.Context, dest string, r io.Reader) error {
	_, err := s.receive(ctx, dest, r)

	return err
}

func (s *impl) ByUUID(ctx context.Context, mi extcmd_model.MountInfo, uuid uuid.UUID) (Summary, error) {
	list, err := s.ListV2(ctx, mi, ListFilter{UUID: uuidext.List{uuid}})
	if err != nil {
		return Summary{}, err
	}

	if len(list) == 0 {
		return Summary{}, fmt.Errorf("%w; subvolume not found by UUID '%s'", errs.ErrNotFound, uuid)
	} else if len(list) == 1 {
		return list[0], nil
	}

	return Summary{}, fmt.Errorf("%w; more than one subvolume match specified UUID '%s'", errs.ErrAmbiguousMatch, uuid)
}

func (s *impl) ListV2(ctx context.Context, mi extcmd_model.MountInfo, filt ListFilter) (res SummaryList, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to list BTRFS subvolumes on mount '%s' using filter '%s'; %w", mi, filt, err)

			s.Log().Err(ctx, err).Send()

			return
		}

		s.Log().Trace(ctx).Msgf("succeeded to list BTRFS subvolumes on mount '%s' using filter '%s', list contains total '%d' item(s)", mi, filt, len(res))
	}()

	if res, err = s.listSubvolumes(ctx, mi); err != nil {
		return nil, err
	}

	filtFn := filt.fn()

	i := 0
	for i < len(res) {
		if !filtFn(res[i]) {
			res = slices.Delete(res, i, i+1)

			continue
		}

		i++
	}

	return
}

func (s *impl) List(ctx context.Context, mi extcmd_model.MountInfo, excludeCreatedAsSnapshot bool) (SummaryList, error) {
	if excludeCreatedAsSnapshot {
		return s.ListV2(ctx, mi, ListFilter{IsSnapshot: new(false)})
	}

	return s.ListV2(ctx, mi, ListFilter{})
}

func (s *impl) ByDirList(ctx context.Context, dir string) (res SummaryList, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to list BTRFS subvolumes on '%s'; %w", dir, err)

			s.Log().Err(ctx, err).Send()

			return
		}

		s.Log().Trace(ctx).Msgf("list of BTRFS subvolumes on '%s' contains total '%d' item(s)", dir, len(res))
	}()

	absPath := dir
	if !filepath.IsAbs(absPath) {
		if absPath, err = filepath.Abs(absPath); err != nil {
			return nil, fmt.Errorf("failed to make path '%s' absolute; %w", dir, err)
		}
	}

	if absPath, err = filepath.EvalSymlinks(absPath); err != nil {
		return nil, fmt.Errorf("failed to evaluate symbolic links of '%s'; %w", dir, err)
	}

	absPath = strings.TrimRight(absPath, string(filepath.Separator))

	stat, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get status of '%s'; %w", absPath, err)
	} else if !stat.IsDir() {
		return nil, fmt.Errorf("path '%s' not points to a directory", dir)
	}

	mi, err := s.findmnt.ByPath(ctx, absPath)
	if err != nil {
		return nil, err
	}

	return s.listSubvolumes(ctx, mi)
}

func (s *impl) SnapshotV2(ctx context.Context, subvol, dest string) (_ Summary, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to make a snapshot of BTRFS subvolume '%s' into '%s'; %w", subvol, dest, err)

			s.Log().Err(ctx, err).Send()

			return
		}

		s.Log().Trace(ctx).Msgf("succeeded to make snapshot of BTRFS subvolume '%s' into '%s'", subvol, dest)
	}()

	if subvol == "" {
		return Summary{}, errors.New("subvolume path argument can't be an empty string")
	}

	if dest == "" {
		return Summary{}, errors.New("snapshot destination path argument can't be an empty string")
	}

	var snapshotPath string

	switch {
	case strings.HasSuffix(dest, string(filepath.Separator)):
		// place snapshot into specified destination and name as subvolume.
		if destStat, destStatErr := os.Stat(dest); destStatErr != nil {
			return Summary{}, fmt.Errorf("failed to get status of '%s'; %w", dest, destStatErr)
		} else if !destStat.IsDir() {
			return Summary{}, fmt.Errorf("destination path '%s' doesn't point to an existing directory", dest)
		}

		snapshotPath = filepath.Join(dest, filepath.Base(subvol))
	default:
		if _, destStatErr := os.Stat(dest); destStatErr == nil {
			return Summary{}, fmt.Errorf("%w; specified snapshot path '%s' already exists; %w;", errs.ErrAlreadyExists, dest, os.ErrExist)
		} else if !os.IsNotExist(destStatErr) {
			return Summary{}, fmt.Errorf("failed to get status of '%s'; %w", dest, destStatErr)
		}

		destDir := filepath.Dir(dest)
		if destStat, destStatErr := os.Stat(destDir); os.IsNotExist(destStatErr) {
			return Summary{}, fmt.Errorf("destination directory '%s' not exists; %w;", destDir, destStatErr)
		} else if destStatErr != nil {
			return Summary{}, fmt.Errorf("failed to get status of '%s'; %w", destDir, destStatErr)
		} else if !destStat.IsDir() {
			return Summary{}, fmt.Errorf("destination path '%s' doesn't point to a directory", destDir)
		}

		snapshotPath = dest
	}

	cmdArgs := []string{subvolSubcmdName, "snapshot", "-r", subvol, snapshotPath}

	_, err = osutil.ExecTextOutputReader(
		ctx,
		osutil.ExecParamsTextOutput{
			Args:              cmdArgs,
			Log:               s.Log(),
			StdoutMaxLines:    extcmd_model.DefaultCommandOutputLines,
			Name:              btrfsCmdName,
			RequiresSuperUser: true,
		},
	)
	if err != nil {
		return Summary{}, err
	}

	res, err := s.showSubvolume(ctx, snapshotPath)
	if err != nil {
		return Summary{}, fmt.Errorf("snapshot was created, but an error occurred while getting info about created snapshot ('%s'); %w", snapshotPath, err)
	}

	return res, nil
}

func (s *impl) IsReadOnly(ctx context.Context, id int64, path string) (readOnly, noChanges bool, err error) {
	var sum Summary

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to get read-only mark for BTRFS subvolume with ID '%d' (path '%s'); %w", id, path, err)

			s.Log().Err(ctx, err).Send()

			return
		}

		s.Log().
			Trace(ctx).
			Msgf("succeeded to get read-only mark for BTRFS subvolume with ID '%d' (path '%s');"+
				" mark '%v', no changes '%v' (creation generation: '%d', generation: '%d', received UUID: '%s')",
				id,
				path,
				readOnly,
				noChanges,
				sum.CreationGeneration,
				sum.Generation,
				sum.ReceivedUUID)
	}()

	cmdArgs := []string{subvolSubcmdName, "show", "-r", stringsext.FormatInt(id), "-b", path}

	r, err := osutil.ExecTextOutputReader(
		ctx,
		osutil.ExecParamsTextOutput{
			Args:              cmdArgs,
			Log:               s.Log(),
			StdoutMaxLines:    extcmd_model.DefaultCommandOutputLines,
			Name:              btrfsCmdName,
			RequiresSuperUser: true,
		},
	)
	if err != nil {
		return false, false, err
	}

	if err = sum.FromBtrfsSubvolumeShowOutputReader(r); err != nil {
		return false, false, err
	}

	if sum.isReadOnly == nil {
		return false, false, errors.New("unexpected miss of read-only mark in subvolume summary object")
	}

	return *sum.isReadOnly,
		sum.CreationGeneration == sum.Generation || sum.ReceivedUUID != uuid.Nil,
		nil
}

func (s *impl) DeleteByIDV2(ctx context.Context, id int64, path string, onlyIfReadOnlyNoChanges bool) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to delete BTRFS subvolume with ID '%d' (path '%s'); %w", id, path, err)

			s.Log().Err(ctx, err).Send()

			return
		}

		s.Log().Trace(ctx).Msgf("succeeded to delete BTRFS subvolume with ID '%d' (path '%s')", id, path)
	}()

	const minSubvolID = 6

	if id < minSubvolID {
		return fmt.Errorf("subvolume ID can't be less than '%d'", minSubvolID)
	}

	if path == "" {
		return errors.New("file system path argument can't be an empty string")
	}

	switch {
	case onlyIfReadOnlyNoChanges:
		var readOnly, noChanges bool

		if readOnly, noChanges, err = s.IsReadOnly(ctx, id, path); err != nil {
			return err
		} else if !readOnly {
			return errors.New("deletion rejected due to specified subvolume is not read-only")
		} else if !noChanges {
			return errors.New("deletion rejected due to specified subvolume was changed since creation")
		}
	}

	cmdArgs := []string{subvolSubcmdName, "delete", "--subvolid", strconv.FormatInt(id, 10), path}

	_, err = osutil.ExecTextOutputReader(
		ctx,
		osutil.ExecParamsTextOutput{
			Args:              cmdArgs,
			Log:               s.Log(),
			StdoutMaxLines:    extcmd_model.DefaultCommandOutputLines,
			Name:              btrfsCmdName,
			RequiresSuperUser: true,
		},
	)

	return err
}

func (s *impl) DeleteByID(ctx context.Context, id int64, path string) (err error) {
	return s.DeleteByIDV2(ctx, id, path, false)
}

func (s *impl) Create(ctx context.Context, path string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create BTRFS subvolume '%s'; %w", path, err)

			s.Log().Err(ctx, err).Send()

			return
		}

		s.Log().Trace(ctx).Msgf("BTRFS subvolume successfully created at '%s'", path)
	}()

	if len(path) == 0 {
		return errors.New("subvolume path argument can't be an empty string")
	}

	cmdArgs := []string{"-q", subvolSubcmdName, "create", "-p", path}

	_, err = osutil.ExecTextOutputReader(
		ctx,
		osutil.ExecParamsTextOutput{
			Args:              cmdArgs,
			Log:               s.Log(),
			StdoutMaxLines:    extcmd_model.DefaultCommandOutputLines,
			Name:              btrfsCmdName,
			RequiresSuperUser: true,
		},
	)

	return
}

func (s *impl) showSubvolume(ctx context.Context, path string) (res Summary, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to get BTRFS subvolume summary for the specified path '%s'; %w", path, err)
		}
	}()

	cmdArgs := []string{subvolSubcmdName, "show", "-b", path}

	r, err := osutil.ExecTextOutputReader(
		ctx,
		osutil.ExecParamsTextOutput{
			Args:              cmdArgs,
			Log:               s.Log(),
			StdoutMaxLines:    extcmd_model.DefaultCommandOutputLines,
			Name:              btrfsCmdName,
			RequiresSuperUser: true,
		},
	)
	if err != nil {
		return Summary{}, err
	}

	if err = res.FromBtrfsSubvolumeShowOutputReader(r); err != nil {
		return
	}

	res.ReachPath = path

	return res, nil
}

func (s *impl) listSubvolumes(ctx context.Context, mi extcmd_model.MountInfo) (SummaryList, error) {
	if !mi.FileSystemType.IsBtrfs() {
		return nil, fmt.Errorf("specified mount info object '%s' not represents BTRFS file system", mi)
	}

	mountSubvolID, ok, err := mi.Options.SubvolIDInt64()
	if err != nil {
		return nil, fmt.Errorf("failed to determine BTRFS mount subvolume ID from mount info '%s' the path belongs to; %w", mi, err)
	} else if !ok {
		return nil, fmt.Errorf("BTRFS mount subvolume ID not determined from mount info '%s' the path belongs to", mi)
	}

	cmdArgs := []string{subvolSubcmdName, "list", "-pcguqR", "--sort=rootid", mi.Target}

	r, err := osutil.ExecTextOutputReader(
		ctx,
		osutil.ExecParamsTextOutput{
			Args:              cmdArgs,
			Log:               s.Log(),
			StdoutMaxLines:    extcmd_model.DefaultCommandOutputLines,
			Name:              btrfsCmdName,
			RequiresSuperUser: true,
		},
	)
	if err != nil {
		return nil, err
	}

	var res SummaryList

	if err = res.FromBtrfsSubvolumeListOutputReader(r); err != nil {
		return nil, err
	}

	s.enrichSubvolReachPath(res, mountSubvolID, mi.Target)
	res.enrichSnapshotsData()

	return res, nil
}

func (*impl) enrichSubvolReachPath(list SummaryList, mountSubvolID int64, mountTarget string) {
	listPointers := make([]*Summary, len(list))

	for i := range len(list) {
		listPointers[i] = &list[i]
	}

	type stackItem struct {
		prefix   string
		subvolID int64
	}

	stack := []stackItem{{subvolID: mountSubvolID, prefix: mountTarget}}

	for len(stack) != 0 {
		item := stack[0]
		stack = slices.Delete(stack, 0, 1)

		j := 0

		for j < len(listPointers) {
			if listPointers[j].ParentID == item.subvolID {
				listPointers[j].ReachPath = filepath.Join(item.prefix, listPointers[j].Path)

				stack = append(stack, stackItem{subvolID: listPointers[j].ID, prefix: listPointers[j].ReachPath})

				listPointers = slices.Delete(listPointers, j, j+1)

				continue
			}

			j++
		}
	}
}

func (s *impl) transmit(ctx context.Context, p TransmitParams) (_ Summary, err error) {
	if err = p.Validate(); err != nil {
		return Summary{}, err
	}

	diffMsg := ""

	if p.SourceParent != "" {
		diffMsg = fmt.Sprintf(" (differential, relatively '%s')", p.SourceParent)
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to execute transmission of BTRFS subvolume '%s'%s into '%s'; %w", p.Source, diffMsg, p.Destination, err)

			s.Log().Err(ctx, err).Send()

			return
		}

		s.Log().
			Info(ctx).
			Msgf("succeeded to execute transmission of BTRFS subvolume '%s'%s into '%s'", p.Source, diffMsg, p.Destination)
	}()

	var (
		sendErr, recvErr    error
		pipeRead, pipeWrite = io.Pipe()
		breakByRecvErr      = errors.New("error occurred on receiver side")
		breakBySendErr      = errors.New("error occurred on sender side")
	)

	defer func() {
		if err != nil {
			return
		}

		if closeErr := pipeRead.Close(); closeErr != nil {
			err = fmt.Errorf("failed to close receive reader; %w", closeErr)
		}

		if closeErr := pipeWrite.Close(); closeErr != nil {
			err = errs.Joinf(err, "failed to close send writer; %w", closeErr)
		}
	}()

	wg := sync.WaitGroup{}

	wg.Go(
		func() {
			if sendErr = s.Send(ctx, p.SourceParent, p.Source, pipeWrite); errors.Is(sendErr, breakByRecvErr) {
				sendErr = nil
			} else if sendErr != nil {
				if closeErr := pipeRead.CloseWithError(breakBySendErr); closeErr != nil {
					sendErr = fmt.Errorf("%w; failed to close receive reader; %w", sendErr, closeErr)
				}

				return
			}

			if closeErr := pipeWrite.Close(); closeErr != nil {
				sendErr = fmt.Errorf("error while closing send writer; %w", err)
			}
		},
	)

	var destRealPath string

	if destRealPath, recvErr = s.receive(ctx, p.Destination, pipeRead); errors.Is(recvErr, breakBySendErr) {
		recvErr = nil
	} else if recvErr != nil {
		if closeErr := pipeWrite.CloseWithError(breakByRecvErr); closeErr != nil {
			recvErr = fmt.Errorf("%w; failed to close send writer; %w", recvErr, closeErr)
		}
	}

	wg.Wait()

	if sendErr != nil {
		err = fmt.Errorf("error while sending; %w", sendErr)
	}

	if recvErr != nil {
		err = errs.Joinf(err, "error while receiving; %w", recvErr)
	}

	if err != nil {
		return Summary{}, err
	}

	destSnapReachPath := filepath.Join(destRealPath, filepath.Base(p.Source))

	destSnap, err := s.showSubvolume(ctx, destSnapReachPath)
	if err != nil {
		return Summary{},
			fmt.Errorf("succeeded to complete receive operation but an error occurred while getting info about received subvolume ('%s'); %w", destSnapReachPath, err)
	}

	return destSnap, nil
}

func (s *impl) receive(ctx context.Context, dest string, r io.Reader) (destRealPath string, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to receive BTRFS subvolume(s) and store it into '%s'; %w", dest, err)

			s.Log().Err(ctx, err).Send()

			return
		}

		s.Log().Trace(ctx).Msgf("succeeded to receive BTRFS subvolume(s) and store it into '%s'", dest)
	}()

	destRealPath, err = filepath.EvalSymlinks(dest)
	if err != nil {
		return "", fmt.Errorf("failed to evaluate real path of '%s'; %w", dest, err)
	}

	destStat, err := os.Stat(destRealPath)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("real destination directory '%s' not exists; %w", destRealPath, err)
	} else if err != nil {
		return "", fmt.Errorf("error while getting status of real destination directory '%s'; %w", destRealPath, err)
	} else if !destStat.IsDir() {
		return "", fmt.Errorf("real destination path '%s' doesn't point to a directory as it expected", destRealPath)
	}

	if !strings.HasSuffix(destRealPath, string(filepath.Separator)) {
		destRealPath = destRealPath + string(filepath.Separator)
	}

	cmdArgs := []string{"receive", "-q", destRealPath}

	return destRealPath,
		osutil.Exec(
			ctx,
			osutil.ExecParamsRawOutput{
				Args:              cmdArgs,
				Log:               s.Log(),
				Name:              btrfsCmdName,
				RequiresSuperUser: true,
				Stdin:             r,
				Stdout:            io.Discard,
			},
		)
}

func New(opts ...OptionFn) (_ Subvol, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", stringsext.TypeStringFor[impl](), err)
		}
	}()

	impl := &impl{}
	if err = fopt.Apply(impl, opts...); err != nil {
		return nil, err
	}

	if impl.findmnt == nil {
		return nil, fmt.Errorf("no service '%s' specified", stringsext.TypeStringForExpr(impl.findmnt))
	}

	return impl, nil
}
