package bridge

import (
	"errors"
	"log"

	"github.com/xuperchain/xuperunion/sql/vfs"
	"github.com/xuperchain/xuperunion/sql/vfs/xfs"
	"github.com/xuperchain/xuperunion/xmodel"
)

type xkv struct {
	ctxmgr *ContextManager
}

func newXfsDB(ctxmgr *ContextManager) xfs.DB {
	return &xkv{
		ctxmgr: ctxmgr,
	}
}

func (x *xkv) Open(name string, pf vfs.ParamFunc) (xfs.Conn, error) {
	var ctxid int64
	pf("ctx", &ctxid)
	ctx, ok := x.ctxmgr.Context(ctxid)
	if !ok {
		return nil, errors.New("context not found")
	}
	return &xkvconn{
		dbname: name,
		ctx:    ctx,
	}, nil
}

type xkvconn struct {
	dbname string
	ctx    *Context
}

func (x *xkvconn) fullkey(key []byte) []byte {
	fullkey := make([]byte, 0, len(x.dbname)+len(key))
	fullkey = append(fullkey, x.dbname...)
	fullkey = append(fullkey, key...)
	return fullkey
}

func (x *xkvconn) Put(key, value []byte) error {
	log.Printf("put ctx %d len %d", x.ctx.ID, len(value))
	return x.ctx.Cache.Put(x.ctx.ContractName, x.fullkey(key), value)
}

func (x *xkvconn) Get(key []byte) ([]byte, error) {
	value, err := x.ctx.Cache.Get(x.ctx.ContractName, x.fullkey(key))
	if err == xmodel.ErrNotFound {
		return nil, xfs.ErrNotFound
	}
	return value.GetPureData().GetValue(), nil
}

func (x *xkvconn) Close() {

}
