package common

import "golang.org/x/exp/constraints"

// Min returns the smaller of two values.
func Min[T constraints.Ordered](x, y T) T {
	if x < y {
		return x
	}

	return y
}

// Max returns the larger of two values.
func Max[T constraints.Ordered](x, y T) T {
	if x > y {
		return x
	}

	return y
}
