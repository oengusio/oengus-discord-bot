package utils

import "log"

func MustNonNil[T interface{}](ptr *T) T {
	if ptr == nil {
		log.Panicln("Cannot dereference nil")
	}

	return *ptr
}

// Map source https://stackoverflow.com/a/71624929
func Map[T, U any](ts []T, f func(T) U) []U {
	us := make([]U, len(ts))
	for i := range ts {
		us[i] = f(ts[i])
	}
	return us
}

func Filter[T any](ss []T, test func(T) bool) (ret []T) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}
