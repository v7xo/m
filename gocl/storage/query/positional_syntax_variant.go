package query

import (
	"strconv"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
)

type PositionalSyntaxVariant int

//nolint:iotamixing
const (
	UnspecifiedPositionalSyntaxVariant = iota
	// Denotes syntax like 'select * from t where t.c1=$1 or t.c2=$1 and t.c3=$2'.
	DollarSignPositionalSyntaxVariant
	// Denotes syntax like 'select * from t where t.c1=? or t.c2=? and t.c3=?'.
	QuestionSignPositionalSyntaxVariant

	DefaultPositionalSyntaxVariant = DollarSignPositionalSyntaxVariant
)

func (v PositionalSyntaxVariant) S() string {
	return v.String()
}

func (v PositionalSyntaxVariant) String() string {
	switch v {
	case UnspecifiedPositionalSyntaxVariant:
		return placeholder.Unspecified
	case DollarSignPositionalSyntaxVariant:
		return "$n"
	case QuestionSignPositionalSyntaxVariant:
		return "?"
	default:
		return strconv.Itoa(int(v))
	}
}
