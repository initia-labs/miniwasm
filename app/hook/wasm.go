package hook

import (
	"context"
	"encoding/json"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

// bridge hook implementation for move
type WasmBridgeHook struct {
	wasmKeeper *wasmkeeper.Keeper
}

func NewWasmBridgeHook(wasmKeeper *wasmkeeper.Keeper) WasmBridgeHook {
	return WasmBridgeHook{wasmKeeper}
}

func (mbh WasmBridgeHook) Hook(ctx context.Context, sender sdk.AccAddress, msgBytes []byte) error {
	var msg wasmtypes.MsgExecuteContract
	decoder := json.NewDecoder(strings.NewReader(string(msgBytes)))
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&msg)
	if err != nil {
		return err
	}

	ms := wasmkeeper.NewMsgServerImpl(mbh.wasmKeeper)
	_, err = ms.ExecuteContract(ctx, &msg)

	return err
}
