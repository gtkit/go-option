// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package example

import (
	"context"
	"crypto/tls"
	"time"
)

// 用于添加 "context" 导包信息，以便在生成代码时判断是否成功去除该无用包信息
var _ context.Context

//go:generate go run ../cmd/go-option/main.go -type GenericExample -with_prefix Example
//go:generate go run ../cmd/go-option/main.go -type GenericExample -output ./generic_example_option.go
//go:generate go run ../cmd/go-option/main.go -type GenericExample -mode append -output ./additional_generic_example_option.go
type GenericExample[T any, U comparable, V ~int] struct {
	A T `opt:"-"`
	B U
	C V
	D string
}

//go:generate go run ../cmd/go-option/main.go -type RedisClientOpt -mode append -output ./generic_example_option.go
type RedisClientOpt struct {
	// Network type to use, either tcp or unix.
	// Default is tcp.
	Network string

	// Redis server address in "host:port" format.
	Addr string

	// Username to authenticate the current connection when Redis ACLs are used.
	// See: https://redis.io/commands/auth.
	Username string

	// Password to authenticate the current connection.
	// See: https://redis.io/commands/auth.
	Password string

	// Redis DB to select after connecting to a server.
	// See: https://redis.io/commands/select.
	DB int

	// Dial timeout for establishing new connections.
	// Default is 5 seconds.
	DialTimeout time.Duration

	// Timeout for socket reads.
	// If timeout is reached, read commands will fail with a timeout error
	// instead of blocking.
	//
	// Use value -1 for no timeout and 0 for default.
	// Default is 3 seconds.
	ReadTimeout time.Duration

	// Timeout for socket writes.
	// If timeout is reached, write commands will fail with a timeout error
	// instead of blocking.
	//
	// Use value -1 for no timeout and 0 for default.
	// Default is ReadTimout.
	WriteTimeout time.Duration

	// Maximum number of socket connections.
	// Default is 10 connections per every CPU as reported by runtime.NumCPU.
	PoolSize int

	// TLS Config used to connect to a server.
	// TLS will be negotiated only if this field is set.
	TLSConfig *tls.Config
}
