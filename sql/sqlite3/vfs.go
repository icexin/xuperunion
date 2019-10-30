package sqlite3

// #include "sqlite3.h"
// #include <stdint.h>
// #include <stdlib.h>
//
// extern int fs_register(char* name, void* appData);
import "C"
import (
	"io"
	"log"
	"unsafe"

	"github.com/xuperchain/xuperunion/sql/vfs"
	"github.com/xuperchain/xuperunion/sql/pointer"
)

// RegisterVFS todo
func RegisterVFS(name string, vfs vfs.FileSystem) {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	C.fs_register(cname, unsafe.Pointer(pointer.Save(vfs)))
}

func genParamFunc(name *C.char) vfs.ParamFunc {
	return func(param string, value interface{}) {
		cparam := C.CString(param)
		defer C.free(unsafe.Pointer(cparam))
		switch v := value.(type) {
		case *int64:
			*v = int64(C.sqlite3_uri_int64(name, cparam, 0))
		case *string:
			*v = C.GoString(C.sqlite3_uri_parameter(name, cparam))
		case *bool:
			ret := C.sqlite3_uri_boolean(name, cparam, 0)
			if ret == 0 {
				*v = false
			} else {
				*v = true
			}
		}
	}
}

//export goFsOpen
func goFsOpen(pfs unsafe.Pointer, name *C.char) unsafe.Pointer {
	fs := pointer.Restore(uintptr(pfs)).(vfs.FileSystem)
	fname := C.GoString(name)
	paramFunc := genParamFunc(name)
	file, err := fs.Open(fname, paramFunc)
	if err != nil {
		log.Printf("open %s error %s", fname, err)
		return nil
	}
	return unsafe.Pointer(pointer.Save(file))
}

//export goFileClose
func goFileClose(pfile unsafe.Pointer) C.int {
	f := pointer.Restore(uintptr(pfile)).(vfs.File)
	f.Close()
	return C.SQLITE_OK
}

func goBytesQuick(pbuf *C.char, n C.int) []byte {
	nn := int(n)
	return (*[1 << 30]byte)(unsafe.Pointer(pbuf))[:nn:nn]
}

//export goFileRead
func goFileRead(pfile unsafe.Pointer, pbuf *C.char, n C.int, offset C.int64_t) C.int {
	f := pointer.Restore(uintptr(pfile)).(vfs.File)
	buf := goBytesQuick(pbuf, n)
	log.Printf("%s read %d at %d", f.Name(), int(n), int(offset))
	nn, err := f.ReadAt(buf, int64(offset))
	if nn == int(n) {
		log.Printf("read full")
		return C.SQLITE_OK
	}
	if err != nil && err != io.EOF {
		log.Printf("read error:%s", err)
		return C.SQLITE_IOERR_READ
	}
	log.Printf("short read:%d remained:%d", nn, len(buf)-nn)
	remain := buf[nn:]
	for i := 0; i < len(remain); i++ {
		remain[i] = 0
	}
	return C.SQLITE_IOERR_SHORT_READ
}

//export goFileWrite
func goFileWrite(pfile unsafe.Pointer, pbuf *C.char, n C.int, offset C.int64_t) C.int {
	f := pointer.Restore(uintptr(pfile)).(vfs.File)
	buf := goBytesQuick(pbuf, n)
	log.Printf("%s write %d at %d", f.Name(), int(n), int(offset))
	_, err := f.WriteAt(buf, int64(offset))
	if err != nil {
		log.Printf("write error:%s", err)
		return C.SQLITE_IOERR
	}
	return C.SQLITE_OK
}

//export goFileSize
func goFileSize(pfile unsafe.Pointer) C.int64_t {
	f := pointer.Restore(uintptr(pfile)).(vfs.File)
	return C.int64_t(f.Size())
}
