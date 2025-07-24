package types

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

func (ac *AccessConfig) ToWasmAccessConfig() *wasmtypes.AccessConfig {
	if ac == nil {
		return nil
	}
	return &wasmtypes.AccessConfig{
		Permission: ac.Permission,
		Addresses:  ac.Addresses,
	}
}
