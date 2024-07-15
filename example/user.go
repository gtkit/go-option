package example

type Embedded struct{}

type Embedded2 struct{}

type Embedded3 struct{}

type Embedded4 struct{}

type Embedded5 struct{}

type Embedded6 struct{}

type Embedded7 struct{}

type Embedded8 struct{}

//go:generate go run ../cmd/go-option/main.go -type User
//go:generate go run ../cmd/go-option/main.go -type User -output ./user_option.go
//go:generate go run ../cmd/go-option/main.go -type User -mode append -output ./additional_user_option.go
type User struct {
	Embedded   `opt:"-"`
	*Embedded2 `opt:"-"`
	E3         Embedded3  `opt:"-"`
	E4         *Embedded4 `opt:"-"`
	Embedded5
	*Embedded6
	E7                 Embedded7
	E8                 *Embedded8
	Username           string
	Email              string
	Address            // combined struct
	ArrayField         [4]int
	SliceField         []int
	MapField           map[string]int
	PtrField           *int
	EmptyStructFiled   struct{}
	SimpleFuncField    func()
	ComplexFuncField   func(a int)
	ComplexFuncFieldV2 func() int
	ComplexFuncFieldV3 func(a int) int
	ComplexFuncFieldV4 func(a int) (int, error)
	ChanField          chan int
	error              // interface
}

type Address struct {
	Street string
	City   string
}
