package xfs

import (
	"encoding/hex"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/xuperchain/xuperunion/sql/vfs"
)

type filekv struct {
}

// NewFileKV todo
func NewFileKV() DB {
	return &filekv{}
}

func (f *filekv) Open(path string, pf vfs.ParamFunc) (Conn, error) {
	err := os.MkdirAll(path, 0755)
	if err != nil {
		return nil, err
	}
	return &fileconn{
		root: path,
	}, nil
}

type fileconn struct {
	root string
}

func (f *fileconn) Put(key, value []byte) error {
	fname := filepath.Join(f.root, hex.EncodeToString(key))
	return ioutil.WriteFile(fname, value, 0644)
}

func (f *fileconn) Get(key []byte) ([]byte, error) {
	fname := filepath.Join(f.root, hex.EncodeToString(key))
	out, err := ioutil.ReadFile(fname)
	if err != nil && os.IsNotExist(err) {
		return nil, ErrNotFound
	}
	return out, err
}

func (f *fileconn) Close() {
}
