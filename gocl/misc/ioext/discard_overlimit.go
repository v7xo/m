package ioext

import (
	"fmt"
	"io"
)

type (
	DiscardOverlimit interface {
		io.Writer

		TotalDiscarded() int
	}

	discardAfterN struct {
		w              io.Writer
		limit          int
		totalDiscarded int
		raiseErr       bool
	}
)

func (d *discardAfterN) Write(p []byte) (n int, err error) {
	if d.limit <= 0 {
		switch {
		case d.raiseErr:
			d.totalDiscarded += len(p)
			err = d.discardError()

			return
		}

		n = len(p)
		d.totalDiscarded += len(p)

		return
	}

	var discardableLen int

	if len(p) > d.limit {
		discardableLen = len(p) - d.limit

		switch {
		case d.raiseErr:
			d.totalDiscarded += discardableLen
			err = d.discardError()

			return
		}

		p = p[:d.limit]
	}

	n, err = d.w.Write(p)

	if n > 0 {
		d.limit -= n
	}

	if n != len(p) {
		return
	}

	n = discardableLen + len(p)
	d.totalDiscarded += discardableLen

	return
}

func (d *discardAfterN) TotalDiscarded() int { return d.totalDiscarded }

func (d *discardAfterN) discardError() error {
	return fmt.Errorf("write to discardable writer failed; total writes reached the limit of size '%d', all subsequent writes will be failed; err: %w", d.limit, io.ErrShortWrite)
}

func NewDiscardOverlimit(lim int, w io.Writer, raiseErr bool) DiscardOverlimit {
	return &discardAfterN{
		limit:    lim,
		w:        w,
		raiseErr: raiseErr,
	}
}
