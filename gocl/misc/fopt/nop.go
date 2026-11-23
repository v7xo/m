package fopt

// [Nop] represents empty functional option.
func Nop[TTarget any](*TTarget) error { return nil }
