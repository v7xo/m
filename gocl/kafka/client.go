package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/segmentio/kafka-go"

	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
)

type clientBuilder struct {
	fopt.LogOpt
	transportContext context.Context
	name             string
	conf             ClientConf
}

func (b *clientBuilder) newClient() (*kafka.Client, error) {
	dialer, err := b.newDialer()
	if err != nil {
		return nil, err
	}

	cli := &kafka.Client{
		Addr:    kafka.TCP(b.conf.Brokers...),
		Timeout: b.conf.RequestTimeout,
		Transport: &kafka.Transport{
			Dial:        dialer.DialFunc,
			DialTimeout: dialer.Timeout,
			ClientID:    dialer.ClientID,
			TLS:         dialer.TLS,
			IdleTimeout: b.conf.IdleTimeout,
			Context:     b.transportContext,
		},
	}

	return cli, nil
}

func (b *clientBuilder) String() string {
	if b == nil {
		return placeholder.Nil
	}

	sb := &strings.Builder{}

	sb.WriteByte('(')

	sb.WriteString(ClientLogScope)

	sb.WriteString(", name: ")
	sb.WriteString(placeholder.UnspecifiedIfEmpty(b.name))

	sb.WriteString(", client ID: ")
	sb.WriteString(placeholder.UnspecifiedIfEmpty(b.conf.ClientID))

	sb.WriteByte(')')

	return sb.String()
}

func (b *clientBuilder) buildTLSConf() (_ *tls.Config, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to buld TLS configuration for '%s'; %w", b, err)

			b.Log().Err(context.Background(), err).Send()
		}
	}()

	tlsConf := &tls.Config{
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: b.conf.TLSInsecure, // nolint:gosec
	}

	if b.conf.TLSCACertPath != "" {
		caPem, readErr := os.ReadFile(b.conf.TLSCACertPath)
		if readErr != nil {
			return nil, readErr
		}

		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(caPem)

		tlsConf.RootCAs = pool
	}

	if b.conf.TLSCertPath == "" {
		return nil, errors.New("no TLS client certificate path specified")
	}

	if b.conf.TLSCertKeyPath == "" {
		return nil, errors.New("no TLS client certificate key path specified")
	}

	cert, err := tls.LoadX509KeyPair(b.conf.TLSCertPath, b.conf.TLSCertKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS client certificate; %w", err)
	}

	tlsConf.Certificates = []tls.Certificate{cert}

	return tlsConf, nil
}

func (b *clientBuilder) newDialer() (_ *kafka.Dialer, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create dialer; %w", err)
		}
	}()

	var tlsConf *tls.Config

	if b.conf.UseTLS {
		switch tlsConf, err = b.buildTLSConf(); err != nil {
		case true:
			return nil, err
		}
	}

	return &kafka.Dialer{
		Timeout:   b.conf.DialTimeout,
		DualStack: true,
		ClientID:  b.conf.ClientID,
		TLS:       tlsConf,
	}, nil
}

func NewClient(opts ...ClientOptionFn) (_ *kafka.Client, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", stringsext.TypeStringFor[*kafka.Client](), err)
		}
	}()

	builder := &clientBuilder{}

	if err = fopt.Apply(builder, opts...); err != nil {
		return
	}

	return builder.newClient()
}

func NewDialer(opts ...ClientOptionFn) (_ *kafka.Dialer, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to create a new instance of '%s'; %w", stringsext.TypeStringFor[*kafka.Dialer](), err)
		}
	}()

	builder := &clientBuilder{}

	if err = fopt.Apply(builder, opts...); err != nil {
		return
	}

	return builder.newDialer()
}
