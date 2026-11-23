package span

import "rn.m.vo7.io/vo7io/m/gocl/value/maybe"

type Of[T byte | int16 | int | int32 | int64 | uint | uint16 | uint32 | uint64 | float32 | float64] struct {
	From maybe.OfOrdered[T]
	To   maybe.OfOrdered[T]
}

func (v *Of[T]) Wide(other Of[T]) Of[T] {
	f := v.From
	t := v.To

	if other.From.NotMiss() &&
		(f.Miss() || other.From.Val() < f.Val()) {
		f = other.From
	}

	if other.To.NotMiss() &&
		(t.Miss() || other.To.Val() > t.Val()) {
		t = other.To
	}

	return Of[T]{
		From: f,
		To:   t,
	}
}

func Val[T byte | int16 | int | int32 | int64 | uint | uint16 | uint32 | uint64 | float32 | float64](from, to T) Of[T] {
	return Of[T]{From: maybe.ValOrdered(from), To: maybe.ValOrdered(to)}
}
