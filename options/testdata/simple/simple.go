package simple

//go:generate go run ../../cmd/go-option/main.go -type User
type User struct {
	Name string `opt:"-"`
	Age  int
	Bio  string
}
