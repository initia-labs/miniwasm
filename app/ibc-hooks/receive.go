package wasm_hooks

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	ibchooks "github.com/initia-labs/initia/x/ibc-hooks"
	nfttransfertypes "github.com/initia-labs/initia/x/ibc/nft-transfer/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

func (h WasmHooks) onRecvIcs20Packet(
	ctx sdk.Context,
	im ibchooks.IBCMiddleware,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
	data transfertypes.FungibleTokenPacketData,
) ibcexported.Acknowledgement {
	isWasmRouted, hookData, err := validateAndParseMemo(data.GetMemo())
	if !isWasmRouted || hookData.Message == nil {
		return im.App.OnRecvPacket(ctx, packet, relayer)
	} else if err != nil {
		return newEmitErrorAcknowledgement(ctx, err)
	}

	msg := hookData.Message
	if allowed, err := h.checkACL(im, ctx, msg.Contract); err != nil {
		return newEmitErrorAcknowledgement(ctx, err)
	} else if !allowed {
		return im.App.OnRecvPacket(ctx, packet, relayer)
	}

	// Validate whether the receiver is correctly specified or not.
	if err := validateReceiver(msg, data.Receiver); err != nil {
		return newEmitErrorAcknowledgement(ctx, err)
	}

	// Calculate the receiver / contract caller based on the packet's channel and sender
	intermediateSender := deriveIntermediateSender(packet.GetDestChannel(), data.GetSender())

	// The funds sent on this packet need to be transferred to the intermediary account for the sender.
	// For this, we override the ICS20 packet's Receiver (essentially hijacking the funds to this new address)
	// and execute the underlying OnRecvPacket() call (which should eventually land on the transfer app's
	// relay.go and send the funds to the intermediary account.
	//
	// If that succeeds, we make the contract call
	data.Receiver = intermediateSender
	bz, err := json.Marshal(data)
	if err != nil {
		return newEmitErrorAcknowledgement(ctx, err)
	}
	packet.Data = bz

	ack := im.App.OnRecvPacket(ctx, packet, relayer)
	if !ack.Success() {
		return ack
	}

	msg.Sender = intermediateSender
	_, err = h.execMsg(ctx, msg)
	if err != nil {
		return newEmitErrorAcknowledgement(ctx, err)
	}

	return ack
}

func (h WasmHooks) onRecvIcs721Packet(
	ctx sdk.Context,
	im ibchooks.IBCMiddleware,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
	data nfttransfertypes.NonFungibleTokenPacketData,
) ibcexported.Acknowledgement {
	isWasmRouted, hookData, err := validateAndParseMemo(data.GetMemo())
	if !isWasmRouted || hookData.Message == nil {
		return im.App.OnRecvPacket(ctx, packet, relayer)
	} else if err != nil {
		return newEmitErrorAcknowledgement(ctx, err)
	}

	msg := hookData.Message
	if allowed, err := h.checkACL(im, ctx, msg.Contract); err != nil {
		return newEmitErrorAcknowledgement(ctx, err)
	} else if !allowed {
		return im.App.OnRecvPacket(ctx, packet, relayer)
	}

	// Validate whether the receiver is correctly specified or not.
	if err := validateReceiver(msg, data.Receiver); err != nil {
		return newEmitErrorAcknowledgement(ctx, err)
	}

	// Calculate the receiver / contract caller based on the packet's channel and sender
	intermediateSender := deriveIntermediateSender(packet.GetDestChannel(), data.GetSender())

	// The funds sent on this packet need to be transferred to the intermediary account for the sender.
	// For this, we override the ICS721 packet's Receiver (essentially hijacking the funds to this new address)
	// and execute the underlying OnRecvPacket() call (which should eventually land on the transfer app's
	// relay.go and send the funds to the intermediary account.
	//
	// If that succeeds, we make the contract call
	data.Receiver = intermediateSender
	bz, err := json.Marshal(data)
	if err != nil {
		return newEmitErrorAcknowledgement(ctx, err)
	}
	packet.Data = bz

	ack := im.App.OnRecvPacket(ctx, packet, relayer)
	if !ack.Success() {
		return ack
	}

	msg.Sender = intermediateSender
	_, err = h.execMsg(ctx, msg)
	if err != nil {
		return newEmitErrorAcknowledgement(ctx, err)
	}

	return ack
}

func (im WasmHooks) execMsg(ctx sdk.Context, msg *wasmtypes.MsgExecuteContract) (*wasmtypes.MsgExecuteContractResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	wasmMsgServer := wasmkeeper.NewMsgServerImpl(im.wasmKeeper)
	return wasmMsgServer.ExecuteContract(ctx, msg)
}
