package cmdhandler

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"rn.m.vo7.io/vo7io/m/gocl/diag/attrname"
	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
	"rn.m.vo7.io/vo7io/m/gocl/value/poly"
)

type Payload interface {
	// IsCmdPayload is a special method representing identity of payload interface type.
	IsCmdPayload()

	Validate() error
	String() string
}

type CmdInput struct {
	Payload       poly.Of[Payload] `json:"payload"`
	CorrelationID string           `json:"-"`
	TraceParent   string           `json:"-"`
}

func (v CmdInput) String() string {
	sb := &strings.Builder{}

	sb.WriteByte('(')

	sb.WriteString("correlation ID: ")
	sb.WriteString(v.CorrelationID)

	sb.WriteString(", payload: ")
	sb.WriteString(v.Payload.String())

	sb.WriteByte(')')

	return sb.String()
}

func (v CmdInput) ToAttrDict() attrname.Dict {
	res := make(attrname.Dict, 2)

	if v.CorrelationID != "" {
		res[attrname.CorrelationID] = v.CorrelationID
	}

	if v.TraceParent != "" {
		res[attrname.TraceParent] = v.TraceParent
	}

	if len(res) == 0 {
		return nil
	}

	return res
}

type cmdHandler interface {
	invoke(ctx context.Context, cmd CmdInput) error
	payloadType() reflect.Type
}

type CmdHandler[T Payload] func(ctx context.Context, p T) error

var _ cmdHandler = CmdHandler[Payload](nil)

func (v CmdHandler[T]) invoke(ctx context.Context, cmd CmdInput) error {
	switch typed := cmd.Payload.Val.(type) {
	case T:
		return v(ctx, typed)
	default:
		return fmt.Errorf("%w; expected '%s', actual '%s'", errs.ErrTypeMismatch, stringsext.TypeStringFor[T](), stringsext.TypeStringOfVal(cmd.Payload.Val))
	}
}

func (v CmdHandler[T]) payloadType() reflect.Type { return reflect.TypeFor[T]() }
