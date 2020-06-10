package evm

import (
	"encoding/hex"
	"fmt"

	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/execution/engine"
	"github.com/hyperledger/burrow/execution/errors"
	"github.com/hyperledger/burrow/execution/evm"
	"github.com/hyperledger/burrow/execution/exec"
	"github.com/xuperchain/xuperchain/core/contract"
	"github.com/xuperchain/xuperchain/core/contract/bridge"
	"github.com/xuperchain/xuperchain/core/contractsdk/go/pb"
)

const (
	initializeMethod = "initialize"
)

type evmCreator struct {
	vm *evm.EVM
}

func newEvmCreator(config *bridge.InstanceCreatorConfig) (bridge.InstanceCreator, error) {
	opt := evm.Options{}
	opt.DebugOpcodes = true
	vm := evm.New(opt)
	return &evmCreator{
		vm: vm,
	}, nil
}

// CreateInstance instances a wasm virtual machine instance which can run a single contract call
func (e *evmCreator) CreateInstance(ctx *bridge.Context, cp bridge.ContractCodeProvider) (bridge.Instance, error) {
	state := newStateManager(ctx)
	return &evmInstance{
		vm:    e.vm,
		ctx:   ctx,
		state: state,
		cp:    cp,
	}, nil
}

func (e *evmCreator) RemoveCache(name string) {
}

type evmInstance struct {
	vm    *evm.EVM
	ctx   *bridge.Context
	state *stateManager
	cp    bridge.ContractCodeProvider
	code  []byte
}

func (e *evmInstance) Exec() error {
	// fmt.Printf("%#v\n", e.ctx)
	var err error
	if e.ctx.Method == initializeMethod {
		code, err := e.cp.GetContractCode(e.ctx.ContractName)
		if err != nil {
			return err
		}
		e.code = code
	} else {
		v, err := e.ctx.Cache.Get("contract", evmCodeKey(e.ctx.ContractName))
		if err != nil {
			fmt.Println("get evm code error")
			return err
		}
		e.code = v.GetPureData().GetValue()
	}
	if e.ctx.Method == initializeMethod {
		return e.deployContract()
	}

	caller, err := ContractAddress(e.state.ctx.Initiator)
	if err != nil {
		return err
	}
	callee, err := ContractAddress(e.ctx.ContractName)
	if err != nil {
		return err
	}
	var gas uint64 = 100000
	input := e.ctx.Args["input"]
	params := engine.CallParams{
		CallType: exec.CallTypeCode,
		Caller:   caller,
		Callee:   callee,
		Input:    input,
		Gas:      &gas,
	}
	out, err := e.vm.Execute(e.state, nil, e, params, e.code)
	if err != nil {
		return err
	}

	e.ctx.Output = &pb.Response{
		Status: 200,
		Body:   []byte(hex.EncodeToString(out)),
	}
	return nil
}

func (e *evmInstance) ResourceUsed() contract.Limits {
	return contract.Limits{}
}

func (e *evmInstance) Release() {
}

func (e *evmInstance) Abort(msg string) {
}

func (e *evmInstance) Call(call *exec.CallEvent, exception *errors.Exception) error {
	return nil
}

func (e *evmInstance) Log(log *exec.LogEvent) error {
	return nil
}

func (e *evmInstance) deployContract() error {
	caller, err := ContractAddress(e.ctx.Initiator)
	if err != nil {
		return err
	}
	var gas uint64 = 100000
	params := engine.CallParams{
		CallType: exec.CallTypeCode,
		Origin:   caller,
		Caller:   caller,
		Callee:   crypto.ZeroAddress,
		Input:    e.ctx.Args["input"],
		Gas:      &gas,
	}
	fmt.Printf("input:%x\n", params.Input)
	contractCode, err := e.vm.Execute(e.state, nil, e, params, e.code)
	if err != nil {
		return err
	}
	key := evmCodeKey(e.ctx.ContractName)
	err = e.ctx.Cache.Put("contract", key, contractCode)
	if err != nil {
		return err
	}
	e.ctx.Output = &pb.Response{
		Status: 200,
	}
	return nil
}

func evmCodeKey(contractName string) []byte {
	return []byte(contractName + "." + "evmcode")
}

func init() {
	bridge.Register(bridge.TypeEvm, "evm", newEvmCreator)
}
