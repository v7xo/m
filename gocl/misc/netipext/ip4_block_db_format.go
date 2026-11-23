package netipext

import "rn.m.vo7.io/vo7io/m/gocl/value/enum"

type IP4BlockDbFormat int16

const (
	IP4BlockDbFormatUnspecified IP4BlockDbFormat = iota
	IP4BlockDbFormatLiteIP2Location1CSV
)

var _ = enum.MustRegister(
	map[IP4BlockDbFormat]string{
		IP4BlockDbFormatUnspecified:         "",
		IP4BlockDbFormatLiteIP2Location1CSV: "lite-ip2location-1-csv",
	},
)

func (v IP4BlockDbFormat) String() string { return enum.String(v) }

func ParseIP4BlockDbFormat(val string) (IP4BlockDbFormat, error) {
	return enum.Parse[IP4BlockDbFormat](val)
}

func MustParseIP4BlockDbFormat(val string) IP4BlockDbFormat {
	res, err := ParseIP4BlockDbFormat(val)
	if err != nil {
		panic(err)
	}

	return res
}
