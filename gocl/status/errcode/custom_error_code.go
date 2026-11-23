package errcode

type CustomErrorCode string

func (v CustomErrorCode) String() string { return string(v) }
