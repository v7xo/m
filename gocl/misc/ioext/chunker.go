package ioext

import (
	"context"
	"errors"
	"fmt"
	iopkg "io"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
)

type Chunk struct {
	Content []byte
	Offset  int64

	// Num represents number of chunk starting from 1.
	Num int64
}

func (c *Chunk) Size() int64 {
	if c == nil {
		return 0
	}

	return int64(len(c.Content))
}

type chunkerOptionFn func(*chunker) error

func ChunkerOption() chunkerOptionFn { return fopt.Nop[chunker] }

func (o chunkerOptionFn) WithLog(val logext.Logger) chunkerOptionFn { return fopt.WithLog(o, val) }

func (o chunkerOptionFn) WithContentLength(val int64) chunkerOptionFn {
	return func(c *chunker) error {
		if err := o(c); err != nil {
			return err
		}

		c.contentLength = val

		return nil
	}
}

func (o chunkerOptionFn) WithReader(val iopkg.ReadCloser) chunkerOptionFn {
	return func(c *chunker) error {
		if err := o(c); err != nil {
			return err
		}

		c.r = val

		return nil
	}
}

func (o chunkerOptionFn) WithChunkLength(val int64) chunkerOptionFn {
	return func(c *chunker) error {
		if err := o(c); err != nil {
			return err
		}

		c.chunkLength = val

		return nil
	}
}

type (
	NextChunkFn func(context.Context) (Chunk, error)

	chunker struct {
		fopt.LogOpt
		r             iopkg.ReadCloser
		buf           []byte
		contentLength int64
		chunkLength   int64
		offset        int64
	}
)

func (s *chunker) NextFn() NextChunkFn {
	if s == nil {
		return nil
	}

	chunkNum := int64(0)

	return func(ctx context.Context) (c Chunk, err error) {
		if c, err = s.next(ctx, chunkNum+1); err != nil {
			return
		}

		chunkNum++

		return
	}
}

func (s *chunker) next(ctx context.Context, num int64) (c Chunk, err error) {
	const errMsgPrefix = "error while calculating next content chunk;"

	if s == nil {
		err = fmt.Errorf("%s chunker is nil", errMsgPrefix)

		return
	}

	if cap(s.buf) == 0 {
		s.buf = make([]byte, 0, s.chunkLength)
	}

	switch s.offset {
	case s.contentLength:
		if closeErr := s.r.Close(); closeErr != nil {
			closeErr = fmt.Errorf("%s failed to close source reader; reader: %T; err: %w", errMsgPrefix, s.r, closeErr)

			s.Log().Err(ctx, closeErr).Send()

			err = closeErr

			return
		}

		err = iopkg.EOF

		return
	}

	bufLen := s.chunkLength
	if s.offset+bufLen > s.contentLength {
		bufLen = s.contentLength - s.offset
	}

	if _, readErr := iopkg.ReadFull(s.r, s.buf[:bufLen]); readErr != nil {
		readErr = fmt.Errorf("%s failed to read next chunk from source reader; reader: %T; err: %s", errMsgPrefix, s.r, readErr.Error())

		s.Log().Err(ctx, readErr).Send()

		err = readErr

		if closeErr := s.r.Close(); closeErr != nil {
			closeErr = fmt.Errorf("%s failed to close source reader; reader: %T; err: %w", errMsgPrefix, s.r, closeErr)

			s.Log().Err(ctx, closeErr).Send()

			err = errors.Join(err, closeErr)
		}

		return
	}

	c = Chunk{
		Content: s.buf[:bufLen],
		Offset:  s.offset,
		Num:     num,
	}

	s.Log().Trace(ctx).Msgf("read next chunk from source reader; offset: %d; len: %d; reader: %T", s.offset, bufLen, s.r)

	s.offset += bufLen

	return
}

func NewChunker(opts ...chunkerOptionFn) (res *chunker, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create instance of chunker; err: %w", err)
		}
	}()

	res = &chunker{chunkLength: DefaultApplicationBufferSize}
	if err = fopt.Apply(res, opts...); err != nil {
		return
	}

	if res.chunkLength < 1 {
		return nil, fmt.Errorf("invalid value of chunk length; value can't be less than 1; value: %d", res.chunkLength)
	}

	if res.contentLength < 1 {
		return nil, fmt.Errorf("invalid value of content length; value can't be less than 1; value: %d", res.chunkLength)
	}

	if res.r == nil {
		return nil, errors.New("no source reader specified")
	}

	if err = res.SetLogScope(chunkerLogScope); err != nil {
		return nil, err
	}

	return
}
