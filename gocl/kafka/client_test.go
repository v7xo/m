//go:build test_kafka

package kafka_test

import (
	"testing"

	"rn.m.vo7.io/vo7io/m/gocl/kafka/kafka_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func Test_NewTopic(t *testing.T) {
	ctx, log := thelper.PreLog(t)
	clientConf := kafka_thelper.FlenvRootClientConf(t)
	createdTopic := kafka_thelper.CreateTopic(t, ctx, clientConf, "topic_*_**")
	log.Info(ctx).Msgf("topic '%s' created", createdTopic)
}
