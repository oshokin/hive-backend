package common

import (
	"crypto/rand"
	"math/big"
	m_rand "math/rand"
	"time"

	"github.com/lazada/awg"
	"golang.org/x/exp/constraints"
)

const defaultTimeout = 1 * time.Minute

// GetRandItemFromList selects a random item from a list and returns it.
// The random number generator used is the secure cryptographic RNG rand.Reader.
// If there are no items in the list, or an error occurs while generating a random number,
// an empty value is returned.
func GetRandItemFromList[T any](list []T) (v T) {
	listLen := int64(len(list))
	if listLen == 0 {
		return
	}

	randBigInt, err := rand.Int(rand.Reader, big.NewInt(listLen))
	if err != nil {
		randBigInt = big.NewInt(m_rand.Int63n(listLen))
	}

	idx := randBigInt.Int64()
	if idx < 0 || idx > listLen-1 {
		return
	}

	v = list[idx]

	return
}

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

// GetDefaultWG initializes default waitgroup and returns it.
func GetDefaultWG(capacity int) *awg.AdvancedWaitGroup {
	wg := &awg.AdvancedWaitGroup{}
	wg.SetCapacity(capacity)
	wg.SetStopOnError(true)
	wg.SetTimeout(defaultTimeout)

	return wg
}
