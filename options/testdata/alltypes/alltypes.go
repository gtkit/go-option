package alltypes

import "io"

type Embedded struct{}

//go:generate go run ../../../cmd/go-option/main.go -type AllTypes
type AllTypes struct {
	Name           string `opt:"-"`
	Age            int
	Ptr            *string
	Slice          []int
	Array          [3]byte
	Map            map[string]int
	Chan           chan int
	SendChan       chan<- int
	RecvChan       <-chan int
	Func           func()
	FuncArgs       func(a int, b string)
	FuncReturn     func() error
	FuncMultiRet   func(x int) (int, error)
	EmptyStruct    struct{}
	Iface          interface{}
	NamedIface     io.Reader
	Embedded
	EmbeddedPtr    *Embedded
	Variadic       func(...int)
}
