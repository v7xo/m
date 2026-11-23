package mapsext

import "errors"

func MustRandKey[TKey comparable, TVal any](m map[TKey]TVal) TKey {
	for k := range m {
		return k
	}

	panic(errors.New("specified map dosn't contain any key"))
}
