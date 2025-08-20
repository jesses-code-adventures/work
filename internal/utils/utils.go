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

func ToPtrNil(t string) *string {
	if t == "" {
		return nil
	}
	return &t
}
