package utils

func ToPtr[T any](t T) *T {
	return &t
}

func FromPtr[T any](t *T) T {
	var zero T
	if t == nil {
		return zero
	}
	return *t
}
