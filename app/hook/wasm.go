package hook

import (
	"context"
	"encoding/json"
	"strings"

	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

// bridge hook implementation for move
type WasmBridgeHook struct {
	ac         address.Codec
	wasmKeeper *wasmkeeper.Keeper
}

func NewWasmBridgeHook(ac address.Codec, wasmKeeper *wasmkeeper.Keeper) WasmBridgeHook {
	return WasmBridgeHook{ac, wasmKeeper}
}

func (mbh WasmBridgeHook) Hook(ctx context.Context, sender sdk.AccAddress, msgBytes []byte) error {
	var msg wasmtypes.MsgExecuteContract
	decoder := json.NewDecoder(strings.NewReader(string(msgBytes)))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&msg)
	if err != nil {
		return err
	}

	// overwrite sender with the actual sender
	msg.Sender, err = mbh.ac.BytesToString(sender)
	if err != nil {
		return err
	}

	ms := wasmkeeper.NewMsgServerImpl(mbh.wasmKeeper)
	_, err = ms.ExecuteContract(ctx, &msg)

	return err
}
