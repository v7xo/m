package kafka_thelper

import (
	"context"
	"regexp"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"

	cl_kafka "rn.m.vo7.io/vo7io/m/gocl/kafka"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/misc/uuid_thelper"
)

func FlenvRootClientConf(t *testing.T) cl_kafka.ClientConf {
	t.Helper()

	conf, err := cl_kafka.FlenvRootClientConf()
	require.NoError(t, err)

	return conf
}

func CreateTopic(t *testing.T, ctx context.Context, clientConf cl_kafka.ClientConf, name string) string {
	t.Helper()

	nameRegexp, err := regexp.Compile(`\*+`)
	require.NoError(t, err)

	name = nameRegexp.ReplaceAllStringFunc(
		name,
		func(s string) string {
			return uuid_thelper.Rand4BHex(t)
		},
	)

	client := NewClient(t, ctx, cl_kafka.ClientOption().Conf(clientConf))

	resp, err := client.CreateTopics(
		ctx,
		&kafka.CreateTopicsRequest{
			Topics: []kafka.TopicConfig{
				{
					Topic:             name,
					NumPartitions:     -1,
					ReplicationFactor: -1,
				},
			},
		},
	)
	require.NoError(t, err, "failed to create Kafka topic '%s'", name)

	for k, err := range resp.Errors {
		require.NoError(t, err, "response on topic creation request (topic name '%s') contains error '%s'", name, k)
	}

	return name
}

func NewClient(t *testing.T, ctx context.Context, opts ...cl_kafka.ClientOptionFn) *kafka.Client {
	t.Helper()

	client, err := cl_kafka.NewClient(fopt.PrependFromContext(ctx, opts...))
	require.NoError(t, err)
	require.NotNil(t, client)

	return client
}
