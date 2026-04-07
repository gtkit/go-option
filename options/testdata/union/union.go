package union

type Number interface {
	~int | ~float64
}

//go:generate go run ../../../cmd/go-option/main.go -type Calc
type Calc[T Number] struct {
	Name  string `opt:"-"`
	Value T
}
