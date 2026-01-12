package wasm_hooks

import (
	"fmt"

	"cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	ibchooks "github.com/initia-labs/initia/x/ibc-hooks"
	ibchookstypes "github.com/initia-labs/initia/x/ibc-hooks/types"
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
	return h.handleOnReceive(ctx, im, packet, relayer, ibchookstypes.ICSData{
		ICS20Data: &data,
	}, func() (sdk.Coins, error) {
		// Extract the denom and amount from the packet data
		localDenom := LocalDenom(packet, data.GetDenom())
		amount, ok := math.NewIntFromString(data.GetAmount())
		if !ok {
			return nil, fmt.Errorf("invalid amount: %s", data.GetAmount())
		}

		// if the denom was migrated, then user will receive L2 denom instead of original IBC denom
		if ok, err := h.opchildKeeper.HasIBCToL2DenomMap(ctx, localDenom); err != nil {
			return nil, err
		} else if ok {
			l2Denom, err := h.opchildKeeper.GetIBCToL2DenomMap(ctx, localDenom)
			if err != nil {
				return nil, err
			}

			// use L2 denom
			localDenom = l2Denom
		}

		return sdk.NewCoins(sdk.NewCoin(localDenom, amount)), nil
	})
}

func (h WasmHooks) onRecvIcs721Packet(
	ctx sdk.Context,
	im ibchooks.IBCMiddleware,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
	data nfttransfertypes.NonFungibleTokenPacketData,
) ibcexported.Acknowledgement {
	return h.handleOnReceive(ctx, im, packet, relayer, ibchookstypes.ICSData{
		ICS721Data: &data,
	}, nil)
}

func (im WasmHooks) execMsg(ctx sdk.Context, msg *wasmtypes.MsgExecuteContract) (*wasmtypes.MsgExecuteContractResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	wasmMsgServer := wasmkeeper.NewMsgServerImpl(im.wasmKeeper)
	return wasmMsgServer.ExecuteContract(ctx, msg)
}

func (h WasmHooks) handleOnReceive(
	ctx sdk.Context,
	im ibchooks.IBCMiddleware,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
	data ibchookstypes.ICSData,
	computeFunds func() (sdk.Coins, error),
) ibcexported.Acknowledgement {
	hookData, routed, err := parseHookData(data.GetMemo())
	if !routed {
		return im.App.OnRecvPacket(ctx, packet, relayer)
	} else if err != nil {
		return newEmitErrorAcknowledgement(err)
	} else if hookData == nil || hookData.Message == nil {
		return im.App.OnRecvPacket(ctx, packet, relayer)
	}

	msg := hookData.Message
	if allowed, err := h.checkACL(im, ctx, msg.Contract); err != nil {
		return newEmitErrorAcknowledgement(err)
	} else if !allowed {
		return newEmitErrorAcknowledgement(fmt.Errorf("contract `%s` is not allowed to be used in ibchooks", msg.Contract))
	}

	// Validate whether the receiver is correctly specified or not.
	if err := validateReceiver(msg, data.GetReceiver()); err != nil {
		return newEmitErrorAcknowledgement(err)
	}

	// Calculate the receiver / contract caller based on the packet's channel and sender
	intermediateSender := DeriveIntermediateSender(packet.GetDestChannel(), data.GetSender())

	// The funds sent on this packet need to be transferred to the intermediary account for the sender.
	// For this, we override the packet receiver (essentially hijacking the funds to this new address)
	// and execute the underlying OnRecvPacket() call (which should eventually land on the transfer app's
	// relay.go and send the funds to the intermediary account.
	//
	// If that succeeds, we make the contract call
	data.SetReceiver(intermediateSender)
	packet.Data = data.GetBytes()

	ack := im.App.OnRecvPacket(ctx, packet, relayer)
	if !ack.Success() {
		return ack
	}

	funds := sdk.NewCoins()
	if computeFunds != nil {
		var err error
		funds, err = computeFunds()
		if err != nil {
			return newEmitErrorAcknowledgement(err)
		}
	}

	msg.Sender = intermediateSender
	msg.Funds = funds
	if _, err := h.execMsg(ctx, msg); err != nil {
		return newEmitErrorAcknowledgement(err)
	}

	return ack
}
