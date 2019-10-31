package sql

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/xuperchain/xuperunion/contract"
	"github.com/xuperchain/xuperunion/contract/bridge"
	"github.com/xuperchain/xuperunion/contractsdk/go/pb"
	"github.com/xuperchain/xuperunion/sql/sqlite3"
)

type executor struct {
}

func NewSQLVM() bridge.Executor {
	return new(executor)
}

func (e *executor) RegisterSyscallService(syscall *bridge.SyscallService) {
}

func (e *executor) NewInstance(ctx *bridge.Context) (bridge.Instance, error) {
	return &instance{
		ctx: ctx,
	}, nil
}

type instance struct {
	ctx *bridge.Context
}

// Exec根据ctx里面的参数执行合约代码
func (i *instance) Exec() error {
	body, err := i.exec()
	if err != nil {
		i.ctx.Output = &pb.Response{
			Status:  500,
			Message: err.Error(),
		}
		return nil
	}
	i.ctx.Output = &pb.Response{
		Status: 200,
		Body:   body,
	}
	return nil
}

func (i *instance) exec() ([]byte, error) {
	dbname, ok := i.ctx.Args["db"]
	if !ok {
		return nil, errors.New("missing db name")
	}
	sqlstr, ok := i.ctx.Args["sql"]
	if !ok {
		return nil, errors.New("missing sql")
	}
	uri := fmt.Sprintf("file:%s?vfs=xfs&ctx=%d", dbname, i.ctx.ID)
	conn, err := sqlite3.Open(uri)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	err = conn.Exec("PRAGMA journal_mode=MEMORY")
	if err != nil {
		return nil, err
	}
	err = conn.Exec("PRAGMA temp_store=MEMORY")
	if err != nil {
		return nil, err
	}

	stmt, err := conn.Prepare(string(sqlstr))
	if err != nil {
		return nil, err
	}
	if stmt == nil {
		return nil, errors.New("nothing to do")
	}
	defer stmt.Close()

	var result []interface{}
	names := stmt.ColumnNames()
	result = append(result, names)
	for {
		hasRow, err := stmt.Step()
		if err != nil {
			return nil, err
		}
		if !hasRow {
			break
		}
		row := make([]interface{}, stmt.ColumnCount())
		for i := 0; i < stmt.ColumnCount(); i++ {
			switch stmt.ColumnType(i) {
			case sqlite3.INTEGER:
				row[i], _, _ = stmt.ColumnInt(i)
			case sqlite3.FLOAT:
				row[i], _, _ = stmt.ColumnDouble(i)
			case sqlite3.TEXT:
				row[i], _, _ = stmt.ColumnText(i)
			case sqlite3.NULL:
				row[i] = nil
			}
		}
		result = append(result, row)
	}
	body, _ := json.Marshal(result)

	return body, nil
}

// ResourceUsed returns the resource used by contract
func (i *instance) ResourceUsed() contract.Limits {
	return contract.Limits{
		Cpu: 1,
	}
}

// Release releases contract instance
func (i *instance) Release() {

}
