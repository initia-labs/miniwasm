package wasm_hooks

import (
	"encoding/json"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	ibchooks "github.com/initia-labs/initia/x/ibc-hooks"
	"github.com/initia-labs/initia/x/ibc-hooks/types"
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
		h.wasmKeeper.Logger(ctx).Error("failed to parse memo", "error", err)
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeHookFailed,
			sdk.NewAttribute(types.AttributeKeyReason, "failed to parse memo"),
			sdk.NewAttribute(types.AttributeKeyError, err.Error()),
		))

		return nil
	}

	// create a new cache context to ignore errors during
	// the execution of the callback
	cacheCtx, write := ctx.CacheContext()

	callback := hookData.AsyncCallback
	if allowed, err := h.checkACL(im, cacheCtx, callback); err != nil {
		h.wasmKeeper.Logger(cacheCtx).Error("failed to check ACL", "error", err)
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeHookFailed,
			sdk.NewAttribute(types.AttributeKeyReason, "failed to check ACL"),
			sdk.NewAttribute(types.AttributeKeyError, err.Error()),
		))

		return nil
	} else if !allowed {
		h.wasmKeeper.Logger(cacheCtx).Error("failed to check ACL", "not allowed")
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeHookFailed,
			sdk.NewAttribute(types.AttributeKeyReason, "failed to check ACL"),
			sdk.NewAttribute(types.AttributeKeyError, "not allowed"),
		))

		return nil
	}

	contractAddr, err := h.ac.StringToBytes(callback)
	if err != nil {
		h.wasmKeeper.Logger(cacheCtx).Error("invalid contract address", "error", err)
		return nil
	}

	success := "false"
	if !isAckError(h.codec, acknowledgement) {
		success = "true"
	}

	// Notify the sender that the ack has been received
	ackAsJson, err := json.Marshal(acknowledgement)
	if err != nil {
		h.wasmKeeper.Logger(cacheCtx).Error("ack is not json object", "error", err)
		return nil
	}

	sudoMsg := []byte(fmt.Sprintf(
		`{"ibc_lifecycle_complete": {"ibc_ack": {"channel": "%s", "sequence": %d, "ack": %s, "success": %s}}}`,
		packet.SourceChannel, packet.Sequence, ackAsJson, success))
	_, err = h.wasmKeeper.Sudo(cacheCtx, contractAddr, sudoMsg)
	if err != nil {
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeHookFailed,
			sdk.NewAttribute(types.AttributeKeyReason, "failed to execute callback"),
			sdk.NewAttribute(types.AttributeKeyError, err.Error()),
		))

		return errorsmod.Wrap(err, "Ack callback error")
	}

	// write the cache context only if the callback execution was successful
	write()

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
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeHookFailed,
			sdk.NewAttribute(types.AttributeKeyReason, "failed to parse memo"),
			sdk.NewAttribute(types.AttributeKeyError, err.Error()),
		))

		h.wasmKeeper.Logger(ctx).Error("failed to parse memo", "error", err)
		return nil
	}

	// create a new cache context to ignore errors during
	// the execution of the callback
	cacheCtx, write := ctx.CacheContext()

	callback := hookData.AsyncCallback
	if allowed, err := h.checkACL(im, cacheCtx, callback); err != nil {
		h.wasmKeeper.Logger(cacheCtx).Error("failed to check ACL", "error", err)
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeHookFailed,
			sdk.NewAttribute(types.AttributeKeyReason, "failed to check ACL"),
			sdk.NewAttribute(types.AttributeKeyError, err.Error()),
		))

		return nil
	} else if !allowed {
		h.wasmKeeper.Logger(cacheCtx).Error("failed to check ACL", "not allowed")
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeHookFailed,
			sdk.NewAttribute(types.AttributeKeyReason, "failed to check ACL"),
			sdk.NewAttribute(types.AttributeKeyError, "not allowed"),
		))

		return nil
	}

	contractAddr, err := h.ac.StringToBytes(callback)
	if err != nil {
		h.wasmKeeper.Logger(cacheCtx).Error("invalid contract address", "error", err)
		return nil
	}

	success := "false"
	if !isAckError(h.codec, acknowledgement) {
		success = "true"
	}

	// Notify the sender that the ack has been received
	ackAsJson, err := json.Marshal(acknowledgement)
	if err != nil {
		h.wasmKeeper.Logger(cacheCtx).Error("ack is not json object", "error", err)
		return nil
	}

	sudoMsg := []byte(fmt.Sprintf(
		`{"ibc_lifecycle_complete": {"ibc_ack": {"channel": "%s", "sequence": %d, "ack": %s, "success": %s}}}`,
		packet.SourceChannel, packet.Sequence, ackAsJson, success))
	_, err = h.wasmKeeper.Sudo(cacheCtx, contractAddr, sudoMsg)
	if err != nil {
		ctx.EventManager().EmitEvent(sdk.NewEvent(
			types.EventTypeHookFailed,
			sdk.NewAttribute(types.AttributeKeyReason, "failed to execute callback"),
			sdk.NewAttribute(types.AttributeKeyError, err.Error()),
		))

		h.wasmKeeper.Logger(cacheCtx).Error("failed to execute callback", "error", err)
		return nil
	}

	// write the cache context only if the callback execution was successful
	write()

	return nil
}
