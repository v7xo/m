package gomigrate

import (
	"bytes"
	"io"
	"os"

	"github.com/golang-migrate/migrate/v4/source"

	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/storagemig/types"
)

type inmemsource struct {
	up   types.SourceVersionList
	down types.SourceVersionList
}

var _ source.Driver = &inmemsource{}

func (s *inmemsource) Open(url string) (source.Driver, error) { return nil, errs.ErrNotSupported }

func (s *inmemsource) Close() error { return nil }

func (s *inmemsource) First() (version uint, err error) {
	if err = s.ensureValid(); err != nil {
		return
	}

	if len(s.up) == 0 {
		err = os.ErrNotExist

		return
	}

	version = s.up[0].Num

	return
}

func (s *inmemsource) Prev(version uint) (prevVersion uint, err error) {
	if err = s.ensureValid(); err != nil {
		return
	}

	for i := range len(s.up) {
		if s.up[i].Num < version && version-s.up[i].Num == 1 {
			prevVersion = s.up[i].Num

			return
		}
	}

	err = os.ErrNotExist

	return
}

func (s *inmemsource) Next(version uint) (nextVersion uint, err error) {
	if err = s.ensureValid(); err != nil {
		return
	}

	for i := range len(s.up) {
		if s.up[i].Num > version && s.up[i].Num-version == 1 {
			nextVersion = s.up[i].Num

			return
		}
	}

	err = os.ErrNotExist

	return
}

func (s *inmemsource) ReadUp(version uint) (r io.ReadCloser, identifier string, err error) {
	if err = s.ensureValid(); err != nil {
		return
	}

	for i := range len(s.up) {
		if s.up[i].Num == version {
			r = io.NopCloser(bytes.NewBufferString(s.up[i].Script))
			identifier = s.up[i].Name

			return
		}
	}

	err = os.ErrNotExist

	return
}

func (s *inmemsource) ReadDown(version uint) (r io.ReadCloser, identifier string, err error) {
	if err = s.ensureValid(); err != nil {
		return
	}

	for i := range len(s.down) {
		if s.down[i].Num == version {
			r = io.NopCloser(bytes.NewBufferString(s.down[i].Script))
			identifier = s.down[i].Name

			return
		}
	}

	err = os.ErrNotExist

	return
}

func (s *inmemsource) ensureValid() error { return verifyUpDownSources(s.up, s.down) }
