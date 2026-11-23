package header

import (
	"fmt"

	"github.com/segmentio/kafka-go"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type Coll map[Name]string

func (coll *Coll) ReadFromKafkaMessage(msg kafka.Message, filt, required Code) (err error) {
	if coll == nil {
		return
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to read headers from Kafka message '(topic: %s, partition: %d, offset: %d)'; %w",
				msg.Topic,
				msg.Partition,
				msg.Offset,
				err)
		}
	}()

	cp := make(Coll, len(msg.Headers))

	for _, h := range msg.Headers {
		n := Name(h.Key)
		c := n.ToCode()

		if filt != CodeUnspecified && filt&c != c {
			continue
		}

		v := string(h.Value)

		if e, ok := cp[n]; ok && len(e) > 0 && len(v) > 0 {
			v = e + "," + v
		} else {
			v = e + v
		}

		cp[n] = v

		required = required &^ c
	}

	if required != CodeUnspecified {
		return fmt.Errorf("%w; no required header(s) '%s' present", errs.ErrDataMapping, required)
	}

	*coll = cp

	return
}

func (coll *Coll) ToKafkaHeader() []kafka.Header {
	if coll == nil {
		return nil
	}

	res := make([]kafka.Header, 0, len(*coll))

	for n, v := range *coll {
		res = append(res, kafka.Header{Key: string(n), Value: []byte(v)})
	}

	return res
}

func (coll *Coll) ToAttrDict() attrname.Dict {
	if coll == nil {
		return nil
	}

	m := make(attrname.Dict, len(*coll))

	for k, v := range *coll {
		if v == "" {
			continue
		}

		n := k.ToAttrName()

		if e, ok := m[n]; ok {
			eStr := e.(string) // nolint

			if len(eStr) > 0 {
				m[n] = eStr + "," + v

				continue
			}
		}

		m[n] = v
	}

	return m
}
