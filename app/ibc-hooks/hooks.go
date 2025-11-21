package wasm_hooks

import (
	"cosmossdk.io/core/address"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	ibchooks "github.com/initia-labs/initia/x/ibc-hooks"
)

var (
	_ ibchooks.OnRecvPacketOverrideHooks            = WasmHooks{}
	_ ibchooks.OnAcknowledgementPacketOverrideHooks = WasmHooks{}
	_ ibchooks.OnTimeoutPacketOverrideHooks         = WasmHooks{}
)

type WasmHooks struct {
	codec         codec.Codec
	ac            address.Codec
	wasmKeeper    *wasmkeeper.Keeper
	opchildKeeper OPChildKeeper
}

func NewWasmHooks(codec codec.Codec, ac address.Codec, wasmKeeper *wasmkeeper.Keeper, opchildKeeper OPChildKeeper) *WasmHooks {
	return &WasmHooks{
		codec:         codec,
		ac:            ac,
		wasmKeeper:    wasmKeeper,
		opchildKeeper: opchildKeeper,
	}
}

func (h WasmHooks) OnRecvPacketOverride(im ibchooks.IBCMiddleware, ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) ibcexported.Acknowledgement {
	if isIcs20, ics20Data := isIcs20Packet(packet.GetData()); isIcs20 {
		return h.onRecvIcs20Packet(ctx, im, packet, relayer, ics20Data)
	}

	if isIcs721, ics721Data := isIcs721Packet(packet.GetData()); isIcs721 {
		return h.onRecvIcs721Packet(ctx, im, packet, relayer, ics721Data)
	}

	return im.App.OnRecvPacket(ctx, packet, relayer)
}

func (h WasmHooks) OnAcknowledgementPacketOverride(im ibchooks.IBCMiddleware, ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
	if isIcs20, ics20Data := isIcs20Packet(packet.GetData()); isIcs20 {
		return h.onAckIcs20Packet(ctx, im, packet, acknowledgement, relayer, ics20Data)
	}

	if isIcs721, ics721Data := isIcs721Packet(packet.GetData()); isIcs721 {
		return h.onAckIcs721Packet(ctx, im, packet, acknowledgement, relayer, ics721Data)
	}

	return im.App.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

func (h WasmHooks) OnTimeoutPacketOverride(im ibchooks.IBCMiddleware, ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	if isIcs20, ics20Data := isIcs20Packet(packet.GetData()); isIcs20 {
		return h.onTimeoutIcs20Packet(ctx, im, packet, relayer, ics20Data)
	}

	if isIcs721, ics721Data := isIcs721Packet(packet.GetData()); isIcs721 {
		return h.onTimeoutIcs721Packet(ctx, im, packet, relayer, ics721Data)
	}

	return im.App.OnTimeoutPacket(ctx, packet, relayer)
}

func (h WasmHooks) checkACL(im ibchooks.IBCMiddleware, ctx sdk.Context, addrStr string) (bool, error) {
	addr, err := h.ac.StringToBytes(addrStr)
	if err != nil {
		return false, err
	}

	return im.HooksKeeper.GetAllowed(ctx, addr)
}
