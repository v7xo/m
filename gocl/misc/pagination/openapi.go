package pagination

import (
	"fmt"

	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

// OpenAPIPageSize размеры страницы для паганации.
type OpenAPIPageSize string

const (
	OpenAPIPageSize100 OpenAPIPageSize = "100"
	OpenAPIPageSize150 OpenAPIPageSize = "150"
	OpenAPIPageSize200 OpenAPIPageSize = "200"
	OpenAPIPageSize250 OpenAPIPageSize = "250"
	OpenAPIPageSize50  OpenAPIPageSize = "50"
)

type OpenAPIParams struct {
	// Limit размеры страницы для паганации
	Limit  OpenAPIPageSize `json:"limit"`
	Offset uint32          `json:"offset"`
}

// nolint:mnd,revive
func (v *Params) FromOpenAPI(val *OpenAPIParams) error {
	if v == nil {
		return nil
	}

	cp := Params{
		Offset: val.Offset,
	}

	switch val.Limit {
	case OpenAPIPageSize50:
		cp.Size = 50
	case OpenAPIPageSize100:
		cp.Size = 100
	case OpenAPIPageSize150:
		cp.Size = 150
	case OpenAPIPageSize200:
		cp.Size = 200
	case OpenAPIPageSize250:
		cp.Size = 250
	case "":
		return fmt.Errorf("failed to map page size enum value specified by OpenAPI pagination model; page size unspecified; err: %w", errs.ErrBadArgs)
	default:
		return fmt.Errorf("failed to map page size enum value specified by OpenAPI pagination model; specified value is unknown; value: %s; err: %w", val.Limit, errs.ErrDataMapping)
	}

	*v = cp

	return nil
}
