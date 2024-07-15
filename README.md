<h1 align="center">
  go-option
</h1>
go-option 是一个在 Go 代码中生成函数选项模式代码的工具。该工具可以根据给定的结构定义自动生成相应的选项代码。
本工具根据 (https://github.com/chenmingyong0423/go-optioner) 修改,生成另一种基于接口实现的函数选项模式代码,供个人学习参考。

---

# 安装
- 1、`go install github.com/gtkit/go-option/cmd/go-option@latest`
- 2、执行 `go-option` 命令检查是否安装成功
```
> go-option
go-option is a tool for generating functional options pattern.
Usage: 
         go-option [flags]
Flags:
         -type <struct name>
         -output <output path>, default: srcDir/opt_xxx_gen.go
``` 
如果你安装成功了，但是提示 `go-option` 命令找不到，请确认是否已将 `$GOPATH/bin` 添加到环境变量中。

# 使用教程
你可以直接使用 `go-option` 命令生成对应结构体的 `functional options` 代码，也可以使用 `go generate` 进行批量生成。
## go-option 命令
- 1、首先，你需要创建一个包含需要生成函数选项模式代码的结构体的 `Go` 文件。在结构体字段中，您可以使用 `opt` 标签来控制是否为 `NewXXX()` 函数的必传参数和生成相应的函数。
```go
package example

type User struct {
	Name   string `opt:"-"`
	Age    int
	Gender string
}

```
如果字段定义了 `opt` 标签，并且值为 `-`，则它将作为 `NewXXX` 函数的必要参数，并且不会生成该字段的 `WithXXX` 函数。

注意：必须声明 `package`。
- 2、在包含结构体定义的文件目录下，执行 `go-option -type XXX` 命令，其中 `XXX` 是结构体的名称。执行命令之后，`go-option` 工具会根据结构体定义生成相应的函数选项模式代码。内容如下所示：
```go
// Generated by go-option -type User; DO NOT EDIT
// If you have any questions, please create issues and submit contributions at:
// https://github.com/gtkit/go-option

package example

type UserInfoOption interface {
	apply(*UserInfo)
}

func NewUser(name string, opts ...UserOption) *User {
	user := &User{
		Name: name,
	}

	for _, opt := range opts {
		opt.apply(userInfo)
	}

	return user
}

type age struct {
	age int
}

func (a age) apply(opt *UserInfo) {
	opt.Age = a.age
}

func WithAge(a int) UserInfoOption {
	return age{age: a}
}

type gender struct {
	gender string
}

func (g gender) apply(opt *UserInfo) {
	opt.Gender = g.gender
}

func WithGender(g string) UserInfoOption {
	return gender{gender: g}
}

```
`go-option` 工具将会生成一个名为 `opt_xxx_gen.go` 的文件，其中 `xxx` 是结构体的名称，例如 `opt_user_gen.go`。该文件包含生成的函数选项代码，用于初始化结构体和设置结构体字段的值。
## go generate 命令
请注意，在执行 `go generate` 命令之前，确保您的项目已经初始化 `Go Modules` 或正确设置了 `GOPATH`，并且您的项目结构符合 `Go Modules` 或 `GOPATH` 的要求。

- 1、首先，你需要创建一个包含需要生成函数选项模式代码的结构体的 `Go` 文件。在结构体定义之上，你需要添加 `//go:generate go-option -type XXX` 的注释，其中 `XXX` 是结构体的名称。这样工具就能根据参数生成相应的代码。在结构体字段中，您可以使用 `opt` 标签来控制是否为 `NewXXX()` 函数的必传参数和生成相应的函数。
```go
package example

//go:generate go-option -type User
type User struct {
	Name   string `opt:"-"`
	Age    int
	Gender string
}
```
如果字段定义了 `opt` 标签，并且值为 `-`，则它将作为 `NewXXX` 函数的必要参数，并且不会生成该字段的 `WithXXX` 函数。

注意：必须声明 `package`。
- 2、在包含结构体定义的文件目录下，执行 `go generate` 命令，这将调用 `go-option` 工具并根据结构体定义生成相应的函数选项模式代码。
