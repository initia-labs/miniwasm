package wasm_hooks

import (
	"encoding/json"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	ibchooks "github.com/initia-labs/initia/x/ibc-hooks"
	nfttransfertypes "github.com/initia-labs/initia/x/ibc/nft-transfer/types"
)

func (h WasmHooks) onAckIcs20Packet(
	ctx sdk.Context,
	im ibchooks.IBCMiddleware,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
	data transfertypes.FungibleTokenPacketData,
) error {
	if err := im.App.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer); err != nil {
		return err
	}

	isWasmRouted, hookData, err := validateAndParseMemo(data.GetMemo())
	if !isWasmRouted || hookData.AsyncCallback == "" {
		return nil
	} else if err != nil {
		return err
	}

	callback := hookData.AsyncCallback
	if allowed, err := h.checkACL(im, ctx, callback); err != nil {
		return err
	} else if !allowed {
		return nil
	}

	contractAddr, err := h.ac.StringToBytes(callback)
	if err != nil {
		return errorsmod.Wrap(err, "Ack callback error")
	}

	success := "false"
	if !isAckError(h.codec, acknowledgement) {
		success = "true"
	}

	// Notify the sender that the ack has been received
	ackAsJson, err := json.Marshal(acknowledgement)
	if err != nil {
		// If the ack is not a json object, error
		return err
	}

	sudoMsg := []byte(fmt.Sprintf(
		`{"ibc_lifecycle_complete": {"ibc_ack": {"channel": "%s", "sequence": %d, "ack": %s, "success": %s}}}`,
		packet.SourceChannel, packet.Sequence, ackAsJson, success))
	_, err = h.wasmKeeper.Sudo(ctx, contractAddr, sudoMsg)
	if err != nil {
		return errorsmod.Wrap(err, "Ack callback error")
	}

	return nil
}

func (h WasmHooks) onAckIcs721Packet(
	ctx sdk.Context,
	im ibchooks.IBCMiddleware,
	packet channeltypes.Packet,
	acknowledgement []byte,
	relayer sdk.AccAddress,
	data nfttransfertypes.NonFungibleTokenPacketData,
) error {
	if err := im.App.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer); err != nil {
		return err
	}

	isWasmRouted, hookData, err := validateAndParseMemo(data.GetMemo())
	if !isWasmRouted || hookData.AsyncCallback == "" {
		return nil
	} else if err != nil {
		return err
	}

	callback := hookData.AsyncCallback
	if allowed, err := h.checkACL(im, ctx, callback); err != nil {
		return err
	} else if !allowed {
		return nil
	}

	contractAddr, err := h.ac.StringToBytes(callback)
	if err != nil {
		return errorsmod.Wrap(err, "Ack callback error")
	}

	success := "false"
	if !isAckError(h.codec, acknowledgement) {
		success = "true"
	}

	// Notify the sender that the ack has been received
	ackAsJson, err := json.Marshal(acknowledgement)
	if err != nil {
		// If the ack is not a json object, error
		return err
	}

	sudoMsg := []byte(fmt.Sprintf(
		`{"ibc_lifecycle_complete": {"ibc_ack": {"channel": "%s", "sequence": %d, "ack": %s, "success": %s}}}`,
		packet.SourceChannel, packet.Sequence, ackAsJson, success))
	_, err = h.wasmKeeper.Sudo(ctx, contractAddr, sudoMsg)
	if err != nil {
		return errorsmod.Wrap(err, "Ack callback error")
	}

	return nil
}
