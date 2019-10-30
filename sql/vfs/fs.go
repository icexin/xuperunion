package vfs

type ParamFunc func(name string, value interface{})

// FileSystem todo
type FileSystem interface {
	Open(name string, pfunc ParamFunc) (File, error)
}

// File todo
type File interface {
	Name() string
	ReadAt(p []byte, offset int64) (int, error)
	WriteAt(p []byte, offset int64) (int, error)
	Size() int64
	Close()
}
