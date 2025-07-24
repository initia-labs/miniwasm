package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

func (msg MsgStoreCodeAdmin) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Creator); err != nil {
		return err
	}

	if err := validateWasmCode(msg.WASMByteCode, wasmtypes.MaxProposalWasmSize); err != nil {
		return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "code bytes %s", err.Error())
	}

	if msg.InstantiatePermission != nil {
		if err := msg.InstantiatePermission.ToWasmAccessConfig().ValidateBasic(); err != nil {
			return errorsmod.Wrap(err, "instantiate permission")
		}
	}
	return nil
}

func validateWasmCode(s []byte, maxSize int) error {
	if len(s) == 0 {
		return errorsmod.Wrap(wasmtypes.ErrEmpty, "is required")
	}
	if len(s) > maxSize {
		return errorsmod.Wrapf(wasmtypes.ErrLimit, "cannot be longer than %d bytes", maxSize)
	}
	return nil
}
