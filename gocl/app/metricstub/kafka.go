package metricstub

import (
	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/diag/prommetric"
)

var (
	KafkaFetchMsgCount = prommetric.MustNewCounterVec(
		false,
		prommetric.CLMetricsSubsystemName("kafka"),
		"fetch_msg_count",
		"indicates the count of messages fetched from Kafka topic by consumer",
		[]string{
			attrname.Topic.String(),
			attrname.ConsumerGroup.String(),
			attrname.HasFailed.String(),
		})

	KafkaSendCommitMsgCount = prommetric.MustNewCounterVec(
		false,
		prommetric.CLMetricsSubsystemName("kafka"),
		"send_commit_msg_count",
		"indicates the count of messages which succeeded/failed to send commit",
		[]string{
			attrname.Topic.String(),
			attrname.ConsumerGroup.String(),
			attrname.HasFailed.String(),
		})

	KafkaProduceMsgCount = prommetric.MustNewCounterVec(
		false,
		prommetric.CLMetricsSubsystemName("kafka"),
		"produce_msg_count",
		"indicates the count of produced messages to Kafka topic",
		[]string{
			attrname.NameAttr.String(),
			attrname.Topic.String(),
			attrname.HasFailed.String(),
		})
)
