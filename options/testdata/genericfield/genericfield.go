package genericfield

// Wrapper uses a generic instantiated type as a field type.
type Pair[K comparable, V any] struct {
	Key   K
	Value V
}

type Wrapper struct {
	Name string        `opt:"-"`
	Data Pair[string, int]
}
