package converters

import "reflect"

func ToPtr[T any](v T) *T {
	return &v
}

// ToPtrOrNil returns a pointer to the given value if it's not the zero value, otherwise nil.
func ToPtrOrNil[T any](value T) *T {
	zero := new(T)
	if reflect.DeepEqual(value, *zero) {
		return nil
	}
	return &value
}

// Deref returns the value of the pointer if it's not nil, otherwise the default (zero) value.
func Deref[T any](v *T) T {
	var def T
	if v == nil {
		return def
	}
	return *v
}
