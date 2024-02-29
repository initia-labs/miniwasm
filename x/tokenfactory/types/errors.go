package types

// DONTCOVER

import (
	fmt "fmt"

	errorsmod "cosmossdk.io/errors"
)

// x/tokenfactory module sentinel errors
var (
	ErrEmptySender              = errorsmod.Register(ModuleName, 2, "empty sender")
	ErrEmptyMintToAddress       = errorsmod.Register(ModuleName, 3, "empty mint-to-address")
	ErrEmptyTransferFromAddress = errorsmod.Register(ModuleName, 4, "empty transfer-from-address")
	ErrEmptyTransferToAddress   = errorsmod.Register(ModuleName, 5, "empty transfer-to-address")
	ErrEmptyNewAdmin            = errorsmod.Register(ModuleName, 6, "empty new-admin")
	ErrDenomExists              = errorsmod.Register(ModuleName, 7, "attempting to create a denom that already exists (has bank metadata)")
	ErrUnauthorized             = errorsmod.Register(ModuleName, 8, "unauthorized account")
	ErrInvalidDenom             = errorsmod.Register(ModuleName, 9, "invalid denom")
	ErrInvalidCreator           = errorsmod.Register(ModuleName, 10, "invalid creator")
	ErrInvalidAuthorityMetadata = errorsmod.Register(ModuleName, 11, "invalid authority metadata")
	ErrInvalidGenesis           = errorsmod.Register(ModuleName, 12, "invalid genesis")
	ErrSubdenomTooLong          = errorsmod.Register(ModuleName, 13, fmt.Sprintf("subdenom too long, max length is %d bytes", MaxSubdenomLength))
	ErrCreatorTooLong           = errorsmod.Register(ModuleName, 14, fmt.Sprintf("creator too long, max length is %d bytes", MaxCreatorLength))
	ErrDenomDoesNotExist        = errorsmod.Register(ModuleName, 15, "denom does not exist")
	ErrBurnFromModuleAccount    = errorsmod.Register(ModuleName, 16, "burning from Module Account is not allowed")
	ErrBeforeSendHookOutOfGas   = errorsmod.Register(ModuleName, 17, "gas meter hit maximum limit")
)
