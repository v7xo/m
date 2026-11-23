package netext_thelper

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/logext_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func RandomTCP4Listen(t *testing.T, ctx context.Context) (ln net.Listener, portNum int) {
	t.Helper()

	log := logext_thelper.FromContext(t, ctx)

	var (
		err        error
		attemptNum int
	)

	randSrc := rand.New(rand.NewSource(time.Now().UnixNano())) // nolint:gosec

loop:
	for attemptNum = 0; attemptNum < rndTCPListenerMaxAttempts; attemptNum++ {
		select {
		case <-ctx.Done():
			err = context.Cause(ctx)

			break loop
		default:
		}

		portNum = int(math.Ceil(randSrc.Float64()*rndTCPListenerPortRangeVolume)) + rndTCPListenerMinPortNum
		log.Trace(ctx).Int(attrname.AddrPort, portNum).Msgf("randomly selected port number to listen on; attempt %d from %d", attemptNum+1, rndTCPListenerMaxAttempts)

		ln, err = net.Listen("tcp4", net.JoinHostPort("", strconv.Itoa(portNum))) //nolint:noctx
		if err == nil || !isErrAddrAlreadyInUse(err) {
			break
		}

		randomDelay := time.Duration(int(math.Ceil(randSrc.Float64()*rndTCPListenerAttemptDelayMaxMs)) * int(time.Millisecond))
		select {
		case <-ctx.Done():
			err = context.Cause(ctx)

			break loop
		case <-time.After(randomDelay):
		}
	}

	require.NoError(t, err, "unexpected error while attempting to listen random TCP port")
	require.NotNil(t, ln, "unexpected nil-result listener")

	lnAddr := ln.Addr().String()

	log.Trace(ctx).
		Int(attrname.AddrPort, portNum).
		Str(attrname.Addr, lnAddr).
		Msgf("listening on randomly selected port number; attempt %d from %d", attemptNum+1, rndTCPListenerMaxAttempts)

	thelper.Cleanup(
		t,
		func() error {
			closeErr := ln.Close()
			if errors.Is(closeErr, net.ErrClosed) {
				return nil
			}

			return closeErr
		},
		"error while closing listener '%s'", lnAddr,
	)

	return
}
