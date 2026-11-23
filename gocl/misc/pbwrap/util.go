package pbwrap

import (
	"errors"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"

	"rn.m.vo7.io/vo7io/m/gocl/misc/ptrof"
)

func Int64Ptr(v *wrapperspb.Int64Value) (res *int64) {
	if v == nil {
		return
	}

	res = &v.Value

	return
}

func Int64Wrap(v *int64) (res *wrapperspb.Int64Value) {
	if v == nil {
		return
	}

	res = wrapperspb.Int64(*v)

	return
}

func Int32Ptr(v *wrapperspb.Int32Value) (res *int32) {
	if v == nil {
		return
	}

	res = &v.Value

	return
}

func Int32Wrap(v *int32) (res *wrapperspb.Int32Value) {
	if v == nil {
		return
	}

	res = wrapperspb.Int32(*v)

	return
}

func BoolPtr(v *wrapperspb.BoolValue) (res *bool) {
	if v == nil {
		return
	}

	res = &v.Value

	return
}

func BoolWrap(v *bool) (res *wrapperspb.BoolValue) {
	if v == nil {
		return
	}

	res = wrapperspb.Bool(*v)

	return
}

func StringPtr(v *wrapperspb.StringValue) (res *string) {
	if v == nil {
		return
	}

	res = &v.Value

	return
}

func StringWrap(v *string) (res *wrapperspb.StringValue) {
	if v == nil {
		return
	}

	res = wrapperspb.String(*v)

	return
}

func BytesPtr(v *wrapperspb.BytesValue) (res []byte) {
	if v == nil {
		return
	}

	res = v.GetValue()

	return
}

func BytesWrap(v []byte) (res *wrapperspb.BytesValue) {
	if len(v) == 0 {
		return
	}

	res = wrapperspb.Bytes(v)

	return
}

func DoublePtr(v *wrapperspb.DoubleValue) (res *float64) {
	if v == nil {
		return
	}

	res = &v.Value

	return
}

func DoubleWrap(v *float64) (res *wrapperspb.DoubleValue) {
	if v == nil {
		return
	}

	res = wrapperspb.Double(*v)

	return
}

func FloatPtr(v *wrapperspb.FloatValue) (res *float32) {
	if v == nil {
		return
	}

	res = &v.Value

	return
}

func FloatWrap(v *float32) (res *wrapperspb.FloatValue) {
	if v == nil {
		return
	}

	res = wrapperspb.Float(*v)

	return
}

func UInt32Ptr(v *wrapperspb.UInt32Value) (res *uint32) {
	if v == nil {
		return
	}

	res = &v.Value

	return
}

func Uint32Wrap(v *uint32) (res *wrapperspb.UInt32Value) {
	if v == nil {
		return
	}

	res = wrapperspb.UInt32(*v)

	return
}

func UInt64Ptr(v *wrapperspb.UInt32Value) (res *uint32) {
	if v == nil {
		return
	}

	res = &v.Value

	return
}

func UInt64Wrap(v *uint64) (res *wrapperspb.UInt64Value) {
	if v == nil {
		return
	}

	res = wrapperspb.UInt64(*v)

	return
}

func TimeUnixPtr(v *timestamppb.Timestamp) (res *int64) {
	if v == nil {
		return
	}

	ts := v.AsTime().Unix()

	res = &ts

	return
}

func TimePtr(v *timestamppb.Timestamp) (res *time.Time, err error) {
	if v == nil {
		return
	} else if !v.IsValid() {
		err = errors.New("specified Protobuf timestamp is not valid")

		return
	}

	res = ptrof.Of(v.AsTime()) // nolint

	return
}

func Time(v *timestamppb.Timestamp) (res time.Time, err error) {
	if v == nil {
		err = errors.New("unable to convert Protobuf timestamp nil value to time")

		return
	} else if !v.IsValid() {
		err = errors.New("specified Protobuf timestamp is not valid")

		return
	}

	res = v.AsTime()

	return
}

func TimeWrap(v *time.Time) (res *timestamppb.Timestamp) {
	if v == nil {
		return
	}

	res = timestamppb.New(*v)

	return
}

func DurationPtr(v *durationpb.Duration) (res *time.Duration) {
	if v == nil {
		return
	}

	dur := v.AsDuration()

	res = &dur

	return
}

func DurationWrap(v *time.Duration) (res *durationpb.Duration) {
	if v == nil {
		return
	}

	res = durationpb.New(*v)

	return
}
