package generic

//go:generate go run ../../cmd/go-option/main.go -type Container
type Container[T any, U comparable] struct {
	ID    string `opt:"-"`
	Value T
	Index U
}
