package utils

import "log"

func MustNonNil[T interface{}](ptr *T) T {
	if ptr == nil {
		log.Panicln("Cannot dereference nil")
	}

	return *ptr
}
