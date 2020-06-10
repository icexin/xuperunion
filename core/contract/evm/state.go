package evm

import (
	"fmt"

	"github.com/hyperledger/burrow/acm"
	"github.com/hyperledger/burrow/binary"
	"github.com/hyperledger/burrow/crypto"
	"github.com/hyperledger/burrow/permission"
	"github.com/xuperchain/xuperchain/core/contract/bridge"
)

type stateManager struct {
	ctx *bridge.Context
}

func newStateManager(ctx *bridge.Context) *stateManager {
	return &stateManager{
		ctx: ctx,
	}
}

// Get an account by its address return nil if it does not exist (which should not be an error)
func (s *stateManager) GetAccount(address crypto.Address) (*acm.Account, error) {
	return &acm.Account{
		Address:     address,
		Balance:     1000,
		Permissions: permission.AllAccountPermissions,
	}, nil
	return nil, nil
}

// Retrieve a 32-byte value stored at key for the account at address, return Zero256 if key does not exist but
// error if address does not
func (s *stateManager) GetStorage(address crypto.Address, key binary.Word256) (value []byte, err error) {
	fmt.Printf("get %s %s\n", address, key)
	v, err := s.ctx.Cache.Get(s.ctx.ContractName, key.Bytes())
	if err != nil {
		return nil, nil
	}
	return v.GetPureData().GetValue(), nil
}

// Updates the fields of updatedAccount by address, creating the account
// if it does not exist
func (s *stateManager) UpdateAccount(updatedAccount *acm.Account) error {
	return nil
}

// Remove the account at address
func (s *stateManager) RemoveAccount(address crypto.Address) error {
	return nil
}

// Store a 32-byte value at key for the account at address, setting to Zero256 removes the key
func (s *stateManager) SetStorage(address crypto.Address, key binary.Word256, value []byte) error {
	fmt.Printf("store %s %s:%x\n", address, key, value)
	return s.ctx.Cache.Put(s.ctx.ContractName, key.Bytes(), value)
}
