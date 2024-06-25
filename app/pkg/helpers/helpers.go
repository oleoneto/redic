package helpers

func PointerTo[T any](value T) *T { return &value }
