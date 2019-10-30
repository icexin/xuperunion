package xfs

import (
	"errors"

	"github.com/xuperchain/xuperunion/sql/vfs"
)

var (
	// ErrNotFound is the error when page not found
	ErrNotFound = errors.New("not found")
)

// DB todo
type DB interface {
	Open(path string, pf vfs.ParamFunc) (Conn, error)
}

// Conn todo
type Conn interface {
	Put(key, value []byte) error
	Get(key []byte) ([]byte, error)
	Close()
}
