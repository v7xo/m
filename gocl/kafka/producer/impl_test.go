//go:build test_kafka

package producer_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"rn.m.vo7.io/vo7io/m/gocl/conf/flenv"
	"rn.m.vo7.io/vo7io/m/gocl/kafka/header"
	"rn.m.vo7.io/vo7io/m/gocl/kafka/kafka_thelper"
	"rn.m.vo7.io/vo7io/m/gocl/kafka/producer"
	"rn.m.vo7.io/vo7io/m/gocl/misc/fopt"
	"rn.m.vo7.io/vo7io/m/gocl/thelper"
)

func Test_Produce(t *testing.T) {
	ctx := thelper.Pre(t)

	conf := producer.ClientConf{}
	err := flenv.PopulatePrefix("", &conf)
	require.NoError(t, err)

	s, err := producer.New(fopt.PrependFromContext(ctx, producer.Option().Conf(conf)))
	require.NoError(t, err)

	jsonPayload := []byte(`{
  "processingDeadline": "2026-04-30T23:59:59+03:00",
  "processingFlags": "NoEPOrderingPublish,NoRepeatProcessingCheck",
  "objectType": "UpdateUser",
  "messageId": "10000000-0000-0000-0000-000000000001",
  "event": "SettingsUpdateRequest",
  "msisdn": "79892090848",
  "items": [
    {
      "alias": "20000000-0000-0000-0000-000000000001",
      "action": "ADD",
      "service": {
        "id": "fbc798ab-5f70-46e5-b5f4-4f75c54dfbfa",
        "serviceSpecification": {
          "id": "ad-blocking",
          "type": "CLIENT_FACING",
          "version": 2
        },
        "product": null,
        "parameters": []
      }
    }
  ],
  "optionalAttributes": {
    "paymentDateShift": {
      "duration": 30,
      "unit": "DAY"
    },
    "productOrderId": "128494a9-67bd-4b52-80be-6772fcbe2685",
    "productSpecificationId": "2db64f1b-33b5-4e5e-8b64-a6857915a113",
    "resultPriceAmount": 0,
    "scheduledModificationEP": false
  },
  "state": "IN_PROGRESS",
  "updated_at": "2026-04-24T12:46:27.86788314Z"
}`)

	hdrs := map[header.Name]string{
		header.NameCorrelationID: "10000000-0000-0000-0000-000000000001",
		header.NameTraceParent:   "nop",
	}

	topic := kafka_thelper.CreateTopic(t, ctx, conf.Client, "test-topic-*")

	err = s.WriteMessage(ctx, topic, []byte("79892090848"), jsonPayload, hdrs)
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)
}
