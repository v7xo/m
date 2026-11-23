package bitsext

import (
	"fmt"
	"strconv"

	"rn.m.vo7.io/vo7io/m/gocl/status/errs"
)

type SegmentSize int

func (s SegmentSize) AsInt() int { return int(s) }

func (s SegmentSize) AsUint64() uint64 { return uint64(s) }

func (s SegmentSize) MaxVal() uint64 {
	return (uint64(1) << s) - 1
}

func (s SegmentSize) String() string { return strconv.FormatInt(int64(s), 10) }

type SegmentsMap []SegmentSize

func (s SegmentsMap) Size() int {
	r := 0

	for i := range len(s) {
		r += int(s[i])
	}

	return r
}

func (s SegmentsMap) Validate() (err error) {
	defer func() {
		if err != nil {
			err = errs.JoinIfNot(fmt.Errorf("invalid bits segments map; %w", err), errs.ErrBadArgs)
		}
	}()

	const segmentsMaxCount = 64

	if len(s) > segmentsMaxCount || len(s) == 0 {
		return fmt.Errorf("count of segments in map can't be greater than '%d' and can't be zero", segmentsMaxCount)
	}

	totalSize := 0

	for i := range len(s) {
		if s[i] < 1 {
			return fmt.Errorf("invalid segment '%s' at '%d' position, size of segment can't be less than 1", s[i], i)
		}

		if totalSize += int(s[i]); totalSize > segmentsMaxCount {
			return fmt.Errorf("total size of map can't be grater than '%d'", segmentsMaxCount)
		}
	}

	if totalSize != segmentsMaxCount {
		return fmt.Errorf("total size of map must be equal to '%d'", segmentsMaxCount)
	}

	return nil
}

func SetSegmentVal[T ~uint64](loc *T, m SegmentsMap, si int, val uint64) error {
	if loc == nil {
		return fmt.Errorf("%w; location pointer is nil", errs.ErrNilValue)
	} else if err := m.Validate(); err != nil {
		return err
	} else if si < 0 || si > len(m)-1 {
		return fmt.Errorf("%w; segment index '%d' is out of range [0, %d)", errs.ErrBadArgs, si, len(m))
	} else if val > m[si].MaxVal() {
		return fmt.Errorf("%w; value '%d' is greater than max. allowed '%d' for segment '%s'", errs.ErrBadArgs, val, m[si].MaxVal(), m[si])
	}

	shift := uint64(0)

	for i := len(m) - 1; i > si; i-- {
		shift += m[i].AsUint64()
	}

	*loc = T(uint64(*loc) | (val << shift))

	return nil
}

func MustSetSegmentVal[T ~uint64](loc *T, m SegmentsMap, si int, val uint64) {
	err := SetSegmentVal(loc, m, si, val)
	if err != nil {
		panic(err)
	}
}

func SegmentVal[T ~uint64](loc T, m SegmentsMap, si int) (uint64, error) {
	if err := m.Validate(); err != nil {
		return 0, err
	} else if si < 0 || si > len(m)-1 {
		return 0, fmt.Errorf("%w; segment index '%d' is out of range [0, %d)", errs.ErrBadArgs, si, len(m))
	}

	shift := uint64(0)

	for i := len(m) - 1; i > si; i-- {
		shift += m[i].AsUint64()
	}

	return (uint64(loc) & (m[si].MaxVal() << shift)) >> shift, nil
}

func MustSegmentVal[T ~uint64](loc T, m SegmentsMap, si int) uint64 {
	res, err := SegmentVal(loc, m, si)
	if err != nil {
		panic(err)
	}

	return res
}
