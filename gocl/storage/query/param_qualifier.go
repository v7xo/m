package query

type ParamQualifier rune

const (
	ColonSignParamQualifier = ':'
	AtSignParamQualifier    = '@'

	DefaultParamQualifier = ColonSignParamQualifier
)
