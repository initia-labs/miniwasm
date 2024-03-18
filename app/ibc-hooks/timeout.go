package wasm_hooks

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	ibchooks "github.com/initia-labs/initia/x/ibc-hooks"
	nfttransfertypes "github.com/initia-labs/initia/x/ibc/nft-transfer/types"
)

func (h WasmHooks) onTimeoutIcs20Packet(
	ctx sdk.Context,
	im ibchooks.IBCMiddleware,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
	data transfertypes.FungibleTokenPacketData,
) error {
	if err := im.App.OnTimeoutPacket(ctx, packet, relayer); err != nil {
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

	sudoMsg := []byte(fmt.Sprintf(
		`{"ibc_lifecycle_complete": {"ibc_timeout": {"channel": "%s", "sequence": %d}}}`,
		packet.SourceChannel, packet.Sequence))
	_, err = h.wasmKeeper.Sudo(ctx, contractAddr, sudoMsg)
	if err != nil {
		return errorsmod.Wrap(err, "Ack callback error")
	}

	return nil
}

func (h WasmHooks) onTimeoutIcs721Packet(
	ctx sdk.Context,
	im ibchooks.IBCMiddleware,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
	data nfttransfertypes.NonFungibleTokenPacketData,
) error {
	if err := im.App.OnTimeoutPacket(ctx, packet, relayer); err != nil {
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

	sudoMsg := []byte(fmt.Sprintf(
		`{"ibc_lifecycle_complete": {"ibc_timeout": {"channel": "%s", "sequence": %d}}}`,
		packet.SourceChannel, packet.Sequence))
	_, err = h.wasmKeeper.Sudo(ctx, contractAddr, sudoMsg)
	if err != nil {
		return errorsmod.Wrap(err, "Ack callback error")
	}

	return nil
}
