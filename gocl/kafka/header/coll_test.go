//go:build test_unit

package header_test

import (
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/kafka/header"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func Test_Coll(t *testing.T) {
	thelper.Pre(t)

	msg := kafka.Message{
		Headers: []kafka.Header{
			{
				Key:   string(header.NameCorrelationID),
				Value: []byte("correlation-id"),
			},
			{
				Key:   string(header.NameTraceParent),
				Value: []byte("trace-parent"),
			},
			{
				Key:   string(header.NameContentType),
				Value: []byte("content-type"),
			},
		},
	}

	// case: required header [messageheader.NameSchemaVersion] is missing.
	c := make(header.Coll)
	err := c.ReadFromKafkaMessage(msg, header.CodeCorrelationID, header.CodeSchemaVersion)
	require.ErrorIs(t, err, errs.ErrDataMapping)

	// case: required header [messageheader.NameSchemaVersion] is missing.
	err = c.ReadFromKafkaMessage(msg, header.CodeCorrelationID, header.CodeSchemaVersion|header.CodeCorrelationID)
	require.ErrorIs(t, err, errs.ErrDataMapping)

	// case: no required headers.
	err = c.ReadFromKafkaMessage(msg, header.CodeCorrelationID, header.CodeUnspecified)
	require.NoError(t, err)
	require.Len(t, c, 1)
	require.Equal(t, "correlation-id", c[header.NameCorrelationID])

	// case: required header [messageheader.NameCorrelationID] present.
	err = c.ReadFromKafkaMessage(msg, header.CodeCorrelationID|header.CodeTraceParent, header.CodeCorrelationID)
	require.NoError(t, err)
	require.Len(t, c, 2)
	require.Equal(t, "correlation-id", c[header.NameCorrelationID])
	require.Equal(t, "trace-parent", c[header.NameTraceParent])
}
