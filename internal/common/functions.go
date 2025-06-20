package common

import (
	"math/rand"
)

// RndElem возвращает случайный элемент из среза
func RndElem[T any](slice []T) T {
	if len(slice) == 0 {
		var zero T
		return zero
	}
	index := rand.Intn(len(slice))
	return slice[index]
}

func IsZero[T comparable](v T) bool {
	var zero T
	return v == zero
}

func Zero[T comparable]() T {
	var zero T
	return zero
}
