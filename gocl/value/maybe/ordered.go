package maybe

import "cmp"

type OfOrdered[T cmp.Ordered] struct{ Of[T] }

func (a OfOrdered[T]) Compare(b OfOrdered[T]) int {
	if !a.notMiss {
		if !b.notMiss {
			return 0
		}

		return 1
	} else if !b.notMiss {
		return -1
	}

	if a.val < b.val {
		return -1
	} else if a.val > b.val {
		return 1
	}

	return 0
}

func ValOrdered[T cmp.Ordered](val T) OfOrdered[T] {
	return OfOrdered[T]{Of[T]{val: val, notMiss: true}}
}

func ValOrderedErr[T cmp.Ordered](val T, err error) (OfOrdered[T], error) {
	if err != nil {
		return OfOrdered[T]{}, err
	}

	return OfOrdered[T]{Of[T]{val: val, notMiss: true}}, nil
}
