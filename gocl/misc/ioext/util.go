package ioext

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"reflect"

	"rn.m.vo7.io/vo7io/m/gocl/diag/logext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
)

// nolint:predeclared
func CopyMaxN(dst io.Writer, src io.Reader, max int64) (written int64, err error) {
	dstDiscard := NewDiscardOverlimit(int(max), dst, true)

	written, err = io.Copy(dstDiscard, src)

	return
}

// SHA256Sum computes SHA-256 hash of content of reader r.
// See also [osutil.FileSHA256Sum].
func SHA256Sum(r io.Reader) ([]byte, error) {
	fBuf := make([]byte, DefaultIOBufferSize)
	h := sha256.New()

	for {
		n, err := io.ReadFull(r, fBuf)
		if err == nil {
			h.Write(fBuf)

			continue
		}

		if errors.Is(err, io.EOF) {
			break
		}

		if errors.Is(err, io.ErrUnexpectedEOF) {
			h.Write(fBuf[:n])

			break
		}

		return nil, err
	}

	return h.Sum(nil), nil
}

// SHA256Sum computes SHA-256 hash of content of reader r and returns HEX-encoded representation of hash.
func SHA256SumHex(r io.Reader) (string, error) {
	h, err := SHA256Sum(r)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h), nil
}

func CloseWithLog[T interface{ Close() error }](ctx context.Context, v T, log logext.Logger) {
	if log == nil || log.IsNop() {
		v.Close()

		return
	}

	err := v.Close()
	if err == nil || errors.Is(err, os.ErrClosed) || errors.Is(err, net.ErrClosed) || errors.Is(err, http.ErrServerClosed) {
		return
	}

	log.LevelFromErr(ctx, err).
		Msgf("error while closing instance of '%s (%s)'; err type: %T",
			stringsext.TypeString(reflect.ValueOf(v).Type()),
			stringsext.TypeStringFor[T](),
			err)
}
