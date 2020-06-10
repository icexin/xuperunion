package evm

import (
	"errors"

	"github.com/btcsuite/btcutil/base58"
	"github.com/hyperledger/burrow/crypto"
	"github.com/xuperchain/xuperchain/core/crypto/hash"
)

func ToEVMAddress(addr string) (crypto.Address, error) {
	rawAddr := base58.Decode(addr)
	if len(rawAddr) < 21 {
		return crypto.ZeroAddress, errors.New("bad address")
	}
	ripemd160Hash := rawAddr[1:21]
	return crypto.AddressFromBytes(ripemd160Hash)
}

func ContractAddress(name string) (crypto.Address, error) {
	rawAddr := hash.UsingRipemd160([]byte(name))
	return crypto.AddressFromBytes(rawAddr)
}
