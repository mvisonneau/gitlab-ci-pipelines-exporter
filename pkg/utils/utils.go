package utils

// Ptr returns a pointer to the given value.
func Ptr[T any](v T) *T {
	return &v
}

// Val returns the actual value from a pointer or a zeroed value if the pointer is nil.
func Val[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}
	return *v
}
