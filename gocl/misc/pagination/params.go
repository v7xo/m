package pagination

import (
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/jackc/pgtype"

	pbt "rn.m.vo7.io/vo7io/m/gocl/pbt/types"
	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

// Params represents generic parameters for pagination.
type Params struct {
	Offset uint32
	Size   uint32
}

func (v Params) String() string {
	return "(offset: " + strconv.FormatUint(uint64(v.Offset), 10) + ", limit: " + strconv.FormatUint(uint64(v.Size), 10) + ")"
}

func Max() Params            { return Params{Offset: 0, Size: math.MaxUint32} }
func OneItem() Params        { return Params{Offset: 0, Size: 1} }
func TwoItem() Params        { return Params{Offset: 0, Size: 2} } // nolint
func Size(val uint32) Params { return Params{Offset: 0, Size: val} }

const (
	// SizeIncreaseStep denotes step in which page size increases.
	// This constant relates to special Protobuf enum type [clpbt.PageSize].
	SizeIncreaseStep uint32 = 50

	// DefaultSize denotes default page size parameter from [Pagination].
	DefaultSize      uint32       = 50
	DefaultSizeProto pbt.PageSize = pbt.PageSize_PAGE_SIZE_50

	Size50       uint32       = 50
	Size50Proto  pbt.PageSize = pbt.PageSize_PAGE_SIZE_50
	Size100      uint32       = 100
	Size100Proto pbt.PageSize = pbt.PageSize_PAGE_SIZE_100
	Size150      uint32       = 150
	Size150Proto pbt.PageSize = pbt.PageSize_PAGE_SIZE_150
	Size200      uint32       = 200
	Size200Proto pbt.PageSize = pbt.PageSize_PAGE_SIZE_200
	Size250      uint32       = 250
	Size250Proto pbt.PageSize = pbt.PageSize_PAGE_SIZE_250
	Size1K       uint32       = 1000
	Size1KProto  pbt.PageSize = pbt.PageSize_PAGE_SIZE_1000
	Size5K       uint32       = 5000
	Size5KProto  pbt.PageSize = pbt.PageSize_PAGE_SIZE_5000
	Size10K      uint32       = 10000
	Size10KProto pbt.PageSize = pbt.PageSize_PAGE_SIZE_10000
)

// FromProto populates this instance from Protobuf [pbt.Pagination].
// nolint:revive
func (v *Params) FromProto(val *pbt.Pagination) error {
	if v == nil {
		return nil
	}

	cp := Params{
		Offset: val.GetOffset(),
	}

	switch val.GetLimit() {
	case pbt.PageSize_PAGE_SIZE_50:
		cp.Size = 50
	case pbt.PageSize_PAGE_SIZE_100:
		cp.Size = 100
	case pbt.PageSize_PAGE_SIZE_150:
		cp.Size = 150
	case pbt.PageSize_PAGE_SIZE_200:
		cp.Size = 200
	case pbt.PageSize_PAGE_SIZE_250:
		cp.Size = 250
	case pbt.PageSize_PAGE_SIZE_1000:
		cp.Size = 1000
	case pbt.PageSize_PAGE_SIZE_5000:
		cp.Size = 5000
	case pbt.PageSize_PAGE_SIZE_10000:
		cp.Size = 10000
	case pbt.PageSize_PAGE_SIZE_UINT32_MAX:
		cp.Size = math.MaxUint32
	case pbt.PageSize_PAGE_SIZE_UNSPECIFIED:
		return fmt.Errorf("failed to map page size enum value specified by Protobuf pagination; page size unspecified; %w", errs.ErrBadArgs)
	default:
		return fmt.Errorf("failed to map page size enum value specified by Protobuf pagination; specified value is unknown; value: %s; %w", val.GetLimit(), errs.ErrDataMapping)
	}

	*v = cp

	return nil
}

func (p Params) MustToProto() *pbt.Pagination {
	res, err := p.ToProto()
	if err != nil {
		panic(err)
	}

	return res
}

// ToProto maps this instance to Protobuf [pbt.Pagination].
func (p *Params) ToProto() (res *pbt.Pagination, err error) {
	if p == nil {
		return
	}

	res = &pbt.Pagination{Offset: p.Offset}

	res.Limit, err = SizeToProto(int(p.Size))

	return
}

func (p *Params) ToPsql() ParamsPsql {
	if p == nil {
		return ParamsPsql{
			Offset: pgtype.Int8{Status: pgtype.Null},
			Limit:  pgtype.Int8{Status: pgtype.Null},
		}
	}

	return ParamsPsql{
		Offset: pgtype.Int8{Status: pgtype.Present, Int: int64(p.Offset)},
		Limit:  pgtype.Int8{Status: pgtype.Present, Int: int64(p.Size)},
	}
}

// FromProto maps Protobuf [pbt.Pagination] to generic pagination parameters ([Params] declared in this package).
func FromProto(val *pbt.Pagination) (Params, error) {
	res := Params{}

	if err := res.FromProto(val); err != nil {
		return Params{}, err
	}

	return res, nil
}

func Proto(size int, offset int) (*pbt.Pagination, error) {
	sizeEnum, err := SizeToProto(size)
	if err != nil {
		return nil, err
	}

	return &pbt.Pagination{
		Offset: uint32(offset),
		Limit:  sizeEnum,
	}, nil
}

// PageSizeProto maps page size to special Protobuf enum [pbt.PageSizeProto].
func SizeToProto(size int) (ps pbt.PageSize, err error) {
	defer func() {
		if err != nil {
			err = errs.BadArgsWrapWith(fmt.Errorf("failed to map specified page size to Protobuf enum; size: %d; %w", size, err))
		}
	}()

	if size < 0 {
		err = errors.New("page size can't be negative")

		return
	} else if size == 0 {
		err = errors.New("zero page size specified")

		return
	} else if size == math.MaxUint32 {
		ps = pbt.PageSize_PAGE_SIZE_UINT32_MAX

		return
	} else if size == int(Size1K) {
		ps = pbt.PageSize_PAGE_SIZE_1000

		return
	} else if size == int(Size5K) {
		ps = pbt.PageSize_PAGE_SIZE_5000

		return
	} else if size == int(Size10K) {
		ps = pbt.PageSize_PAGE_SIZE_10000

		return
	}

	ps = pbt.PageSize(uint32(size) / SizeIncreaseStep)

	switch ps {
	case
		pbt.PageSize_PAGE_SIZE_50,
		pbt.PageSize_PAGE_SIZE_100,
		pbt.PageSize_PAGE_SIZE_150,
		pbt.PageSize_PAGE_SIZE_200,
		pbt.PageSize_PAGE_SIZE_250:
		break
	case
		pbt.PageSize_PAGE_SIZE_1000,
		pbt.PageSize_PAGE_SIZE_10000,
		pbt.PageSize_PAGE_SIZE_5000,
		pbt.PageSize_PAGE_SIZE_UINT32_MAX,
		pbt.PageSize_PAGE_SIZE_UNSPECIFIED:
		fallthrough
	default:
		err = errors.New("specified page size is out of range")
	}

	if err != nil {
		return
	} else if size != int(uint32(ps)*SizeIncreaseStep) {
		err = errors.New("specified page size not aligned to one of valid Protobuf values")

		return
	}

	return
}

// nolint:revive
func SizeFromProto(ps pbt.PageSize) (size uint32, err error) {
	defer func() {
		if err != nil {
			err = errs.BadArgsWrapWith(fmt.Errorf("failed to map specified Protobuf page size enum; value: %d; %w", ps, err))
		}
	}()

	switch ps {
	case pbt.PageSize_PAGE_SIZE_50:
		size = Size50
	case pbt.PageSize_PAGE_SIZE_100:
		size = Size100
	case pbt.PageSize_PAGE_SIZE_150:
		size = Size150
	case pbt.PageSize_PAGE_SIZE_200:
		size = Size200
	case pbt.PageSize_PAGE_SIZE_250:
		size = Size250
	case pbt.PageSize_PAGE_SIZE_1000:
		size = Size1K
	case pbt.PageSize_PAGE_SIZE_5000:
		size = Size5K
	case pbt.PageSize_PAGE_SIZE_10000:
		size = Size10K
	case pbt.PageSize_PAGE_SIZE_UINT32_MAX:
		size = math.MaxUint32
	case pbt.PageSize_PAGE_SIZE_UNSPECIFIED:
		err = errors.New("value unspecified")
	default:
		err = fmt.Errorf("specified value is unknown; value: %s; %w", ps, errs.ErrDataMapping)
	}

	return
}

func DefaultProto() *pbt.Pagination {
	return &pbt.Pagination{
		Offset: 0,
		Limit:  DefaultSizeProto,
	}
}

type ParamsPsql struct {
	Offset pgtype.Int8 `db:"offset"`
	Limit  pgtype.Int8 `db:"limit"`
}
