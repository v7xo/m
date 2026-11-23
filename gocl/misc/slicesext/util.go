package slicesext

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"strconv"

	"golang.org/x/sync/errgroup"

	"rn.m.vo7.io/vo7io/m/gocl/misc/syncext"
)

// Shuf pseudo-randomizes the order of slice S elements.
func Shuf[S ~[]E, E any](src S) S {
	// nolint:mnd
	if len(src) < 2 {
		return src
	}

	for range len(src) / 2 {
		a, b := rand.IntN(len(src)), rand.IntN(len(src)) // nolint:gosec
		src[a], src[b] = src[b], src[a]
	}

	return src
}

type Segment[S ~[]E, E any] struct {
	// S is slice representing segment of source slice.
	S S
	// Offset points to start of this segment in source slice.
	Offset int
	// I is a zero-based index of this segment.
	I int
}

func (v Segment[S, E]) String() string {
	return "(offset: " + strconv.FormatInt(int64(v.Offset), 10) + ", size: " + strconv.FormatInt(int64(len(v.S)), 10)
}

func Split[S ~[]E, E any](s S, size int) []Segment[S, E] {
	if size < 1 {
		panic(fmt.Errorf("invalid value of segment size, can't be less than 1; value: %d", size))
	}

	if s == nil {
		return nil
	} else if len(s) == 0 {
		return []Segment[S, E]{}
	} else if len(s) <= size {
		return []Segment[S, E]{{S: s}}
	}

	res := make([]Segment[S, E], 0, len(s)/size+1)

	lo := 0
	i := 0

	for lo < len(s) {
		hi := min(lo+size, len(s))
		seg := s[lo:hi:hi]

		res = append(res, Segment[S, E]{S: seg, I: i, Offset: lo})

		lo += len(seg)
		i++
	}

	return res
}

type ProcPartitionFn[S ~[]E, E any] func(ctx context.Context, partIndex int, part S) error

func ProcPartitions[S ~[]E, E any](ctx context.Context, s S, partitionSize int, maxDOP syncext.MaxDOP, fn ProcPartitionFn[S, E]) error {
	if len(s) == 0 {
		return nil
	}

	dop, err := maxDOP.Effective()
	if err != nil {
		return err
	}

	if partitionSize < 1 {
		return fmt.Errorf("invalid value of partition size - can't be less than 1; value: %d", partitionSize)
	} else if fn == nil {
		return errors.New("no processor function specified")
	} else if maxDOP < 0 {
		return fmt.Errorf("invalid value of max. DOP - can't be less than 0; value: %d", maxDOP)
	}

	if len(s) <= partitionSize {
		return fn(ctx, 0, s)
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.SetLimit(dop)

	lo := 0
	partitionIndex := 0

	for lo < len(s) {
		hi := min(lo+partitionSize, len(s))
		partition := s[lo:hi:hi]
		index := partitionIndex

		eg.Go(func() error { return fn(ctx, index, partition) })

		lo += len(partition)
		partitionIndex++
	}

	return eg.Wait()
}

func Uniq[S ~[]E, E comparable](s S) (res S) {
	if s == nil {
		return
	}

	dict := make(map[E]struct{}, min(1000, len(s))) // nolint:mnd,revive
	res = make(S, 0, min(1000, len(s)))             // nolint:mnd,revive

	for i := range len(s) {
		if _, ok := dict[s[i]]; ok {
			continue
		}

		dict[s[i]] = struct{}{}

		res = append(res, s[i])
	}

	return
}

func SelectUniq[S ~[]E, E any, V comparable](s S, sel func(e E) (v []V, skip bool)) (res []V) {
	if s == nil {
		return
	} else if len(s) == 0 {
		res = make([]V, 0)

		return
	}

	res = make([]V, 0, len(s))
	dict := make(map[V]struct{}, len(s))

	for i := range s {
		vals, skip := sel(s[i])
		if skip {
			continue
		}

		for j := range vals {
			if _, ok := dict[vals[j]]; ok {
				continue
			}

			dict[vals[j]] = struct{}{}

			res = append(res, vals[j])
		}
	}

	return
}

func Select[S ~[]E, E any, V comparable](s S, sel func(e E) (v V, skip bool), uniq bool) (res []V) {
	if s == nil {
		return
	} else if len(s) == 0 {
		res = make([]V, 0)

		return
	}

	res = make([]V, 0, len(s))

	var dict map[V]struct{}

	for i := range s {
		val, skip := sel(s[i])
		if skip {
			continue
		}

		switch {
		case uniq:
			if dict == nil {
				dict = make(map[V]struct{}, len(s))
			}

			if _, ok := dict[val]; ok {
				continue
			}

			dict[val] = struct{}{}
		}

		res = append(res, val)
	}

	return
}

func Last[T any](s []T) (res T, ok bool) {
	if len(s) == 0 {
		return
	}

	res, ok = s[len(s)-1], true

	return
}

func LastNonDefault[T comparable](s []T) (res T, ok bool) {
	var defaultValue T

	for i := len(s) - 1; i > -1; i-- {
		if s[i] != defaultValue {
			res = s[i]
			ok = true

			return
		}
	}

	return
}

func Empty[S ~[]E, E any](s *S) bool { return s != nil && *s != nil && len(*s) == 0 }

type Grouping[K comparable, S ~[]E, E any] map[K]S

func GroupSkip[S ~[]E, E any, K comparable](s S, sel func(e E) (k K, skip bool)) (res Grouping[K, S, E]) {
	if s == nil {
		return
	} else if len(s) == 0 {
		res = make(Grouping[K, S, E], 0)

		return
	}

	res = make(Grouping[K, S, E], len(s))

	for i := range s {
		key, skip := sel(s[i])
		if skip {
			continue
		}

		forKey := res[key]
		forKey = append(forKey, s[i])
		res[key] = forKey
	}

	return
}

func Group[S ~[]E, E any, K comparable](s S, sel func(e E) K) (res Grouping[K, S, E]) {
	if s == nil {
		return
	} else if len(s) == 0 {
		res = make(Grouping[K, S, E], 0)

		return
	}

	res = make(Grouping[K, S, E], len(s))

	for i := range s {
		key := sel(s[i])
		res[key] = append(res[key], s[i])
	}

	return
}
