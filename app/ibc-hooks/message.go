package wasm_hooks

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

const (
	// The memo key is used to parse ics-20 or ics-712 memo fields.
	wasmHookMemoKey = "wasm"
)

// HookData defines a wrapper for wasm execute message
// and async callback.
type HookData struct {
	// Message is a wasm execute message which will be executed
	// at `OnRecvPacket` of receiver chain.
	Message *wasmtypes.MsgExecuteContract `json:"message,omitempty"`

	// AsyncCallback is a contract address
	AsyncCallback string `json:"async_callback,omitempty"`
}
