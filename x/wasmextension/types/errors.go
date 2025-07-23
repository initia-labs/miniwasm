package types

import (
	errorsmod "cosmossdk.io/errors"
)

// Codes for wasm contract errors
var (
	DefaultCodespace = ModuleName

	// Note: never use code 1 for any errors - that is reserved for ErrInternal in the core cosmos sdk

	// ErrCreateFailed error for wasm code that has already been uploaded or failed
	ErrUnauthorized = errorsmod.Register(DefaultCodespace, 2, "unauthorized")
)
