package complex

import "time"

type Embedded struct{}

//go:generate go run ../../cmd/go-option/main.go -type Server
type Server struct {
	Addr        string        `opt:"-"`
	Port        int           `opt:"-"`
	Timeout     time.Duration
	MaxConns    int
	TLSEnabled  bool
	Tags        []string
	Metadata    map[string]string
	Handler     func() error
	Embedded
}
