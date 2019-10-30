package xfs

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/xuperchain/xuperunion/sql/vfs"
)

type ldb struct {
}

// NewLdbKV todo
func NewLdbKV() DB {
	return new(ldb)
}

func (l *ldb) Open(path string, pf vfs.ParamFunc) (Conn, error) {
	conn, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return &ldbconn{
		conn: conn,
	}, nil
}

type ldbconn struct {
	conn *leveldb.DB
}

func (l *ldbconn) Put(key, value []byte) error {
	return l.conn.Put(key, value, nil)
}

func (l *ldbconn) Get(key []byte) ([]byte, error) {
	out, err := l.conn.Get(key, nil)
	if err != nil && err == leveldb.ErrNotFound {
		return nil, ErrNotFound
	}
	return out, err
}

func (l *ldbconn) Close() {
	l.conn.Close()
}
