package xfs

import (
	"encoding/binary"
	"io"
	"strconv"

	"github.com/xuperchain/xuperunion/sql/vfs"
)

const (
	pageSize = 4096
)

var (
	metakey = []byte{0, 0, 0, 0}
)

type filesystem struct {
	db DB
}

// NewFileSystem todo
func NewFileSystem(db DB) vfs.FileSystem {
	return &filesystem{
		db: db,
	}
}

func (fs *filesystem) Open(name string, pf vfs.ParamFunc) (vfs.File, error) {
	conn, err := fs.db.Open(name, pf)
	if err != nil {
		return nil, err
	}
	f := &file{
		name: name,
		conn: conn,
	}
	err = f.init()
	if err != nil {
		return nil, err
	}
	return f, nil
}

type file struct {
	name string
	size int64
	conn Conn
}

func (f *file) init() error {
	f.size = f.readFileSize()
	return nil
}

func (f *file) Name() string {
	return f.name
}

func (f *file) ReadAt(p []byte, offset int64) (int, error) {
	total := int64(len(p))
	nread := int64(0)
	for {
		no := (offset+nread)/pageSize + 1
		off := (offset + nread) % pageSize
		key := make([]byte, 4)
		binary.BigEndian.PutUint32(key, uint32(no))
		buf, err := f.conn.Get(key)
		if err != nil {
			if err == ErrNotFound {
				return int(nread), io.EOF
			}
			return int(nread), err
		}
		content := buf[off:]
		n := copy(p[nread:], content)
		nread += int64(n)
		if nread >= total {
			return int(nread), nil
		}
	}
}

func (f *file) WriteAt(p []byte, offset int64) (int, error) {
	var err error
	total := int64(len(p))
	nwrite := int64(0)
	for {
		no := (offset+nwrite)/pageSize + 1
		buf := p[nwrite : nwrite+pageSize]
		key := make([]byte, 4)
		binary.BigEndian.PutUint32(key, uint32(no))
		err = f.conn.Put(key, buf)
		if err != nil {
			break
		}
		nwrite += pageSize
		if nwrite >= total {
			break
		}
	}
	if nwrite != 0 && offset+nwrite > f.size {
		f.updateFileSize(offset + nwrite)
	}
	return int(nwrite), err
}

func (f *file) Size() int64 {
	return f.size
}

func (f *file) Close() {
	f.conn.Close()
}

func (f *file) readFileSize() int64 {
	value, err := f.conn.Get(metakey)
	if err != nil {
		return 0
	}
	size, _ := strconv.Atoi(string(value))
	return int64(size)
}

func (f *file) updateFileSize(size int64) {
	f.conn.Put(metakey, []byte(strconv.Itoa(int(size))))
	f.size = size
}
