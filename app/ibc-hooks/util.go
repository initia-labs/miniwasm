package wasm_hooks

import (
	"encoding/json"
	"fmt"
	"strings"

	"cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	nfttransfertypes "github.com/initia-labs/initia/x/ibc/nft-transfer/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

const senderPrefix = "ibc-wasm-hook-intermediary"

// DeriveIntermediateSender compute intermediate sender address
// Bech32(Hash(Hash("ibc-hook-intermediary") + channelID/sender))
func DeriveIntermediateSender(channel, originalSender string) string {
	senderStr := fmt.Sprintf("%s/%s", channel, originalSender)
	senderAddr := sdk.AccAddress(address.Hash(senderPrefix, []byte(senderStr)))
	return senderAddr.String()
}

func isIcs20Packet(packetData []byte) (isIcs20 bool, ics20data transfertypes.FungibleTokenPacketData) {
	var data transfertypes.FungibleTokenPacketData
	decoder := json.NewDecoder(strings.NewReader(string(packetData)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&data); err != nil {
		return false, data
	}
	return true, data
}

const wasmPortPrefix = "wasm."

func isIcs721Packet(packetData []byte) (isIcs721 bool, ics721data nfttransfertypes.NonFungibleTokenPacketData) {
	// Use wasm port prefix to ack like normal wasm chain.
	//
	// initia l1 is handling encoding and decoding depends on port id,
	// so miniwasm should ack like normal wasm chain.
	if data, err := nfttransfertypes.DecodePacketData(packetData, wasmPortPrefix); err != nil {
		return false, data
	} else {
		return true, data
	}
}

func validateAndParseMemo(memo string) (
	isWasmRouted bool,
	hookData HookData,
	err error,
) {
	isWasmRouted, metadata := jsonStringHasKey(memo, wasmHookMemoKey)
	if !isWasmRouted {
		return
	}

	wasmHookRaw := metadata[wasmHookMemoKey]

	// parse wasm raw bytes to execute message
	bz, err := json.Marshal(wasmHookRaw)
	if err != nil {
		err = errors.Wrap(channeltypes.ErrInvalidPacket, err.Error())
		return
	}

	err = json.Unmarshal(bz, &hookData)
	if err != nil {
		err = errors.Wrap(channeltypes.ErrInvalidPacket, err.Error())
		return
	}

	return
}

func validateReceiver(msg *wasmtypes.MsgExecuteContract, receiver string) error {
	if receiver != msg.Contract {
		return errors.Wrap(channeltypes.ErrInvalidPacket, "receiver is not properly set")
	}

	return nil
}

// jsonStringHasKey parses the memo as a json object and checks if it contains the key.
func jsonStringHasKey(memo, key string) (found bool, jsonObject map[string]interface{}) {
	jsonObject = make(map[string]interface{})

	// If there is no memo, the packet was either sent with an earlier version of IBC, or the memo was
	// intentionally left blank. Nothing to do here. Ignore the packet and pass it down the stack.
	if len(memo) == 0 {
		return false, jsonObject
	}

	// the jsonObject must be a valid JSON object
	err := json.Unmarshal([]byte(memo), &jsonObject)
	if err != nil {
		return false, jsonObject
	}

	// If the key doesn't exist, there's nothing to do on this hook. Continue by passing the packet
	// down the stack
	_, ok := jsonObject[key]
	if !ok {
		return false, jsonObject
	}

	return true, jsonObject
}

// newEmitErrorAcknowledgement creates a new error acknowledgement after having emitted an event with the
// details of the error.
func newEmitErrorAcknowledgement(err error) channeltypes.Acknowledgement {
	return channeltypes.Acknowledgement{
		Response: &channeltypes.Acknowledgement_Error{
			Error: fmt.Sprintf("ibc wasm hook error: %s", err.Error()),
		},
	}
}

// isAckError checks an IBC acknowledgement to see if it's an error.
// This is a replacement for ack.Success() which is currently not working on some circumstances
func isAckError(appCodec codec.Codec, acknowledgement []byte) bool {
	var ack channeltypes.Acknowledgement
	if err := appCodec.UnmarshalJSON(acknowledgement, &ack); err == nil && !ack.Success() {
		return true
	}

	return false
}

// MustExtractDenomFromPacketOnRecv takes a packet with a valid ICS20 token data in the Data field and returns the
// denom as represented in the local chain.
// If the data cannot be unmarshalled this function will panic
func MustExtractDenomFromPacketOnRecv(packet ibcexported.PacketI) string {
	var data transfertypes.FungibleTokenPacketData
	if err := json.Unmarshal(packet.GetData(), &data); err != nil {
		panic("unable to unmarshal ICS20 packet data")
	}

	var denom string
	if transfertypes.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), data.Denom) {
		// remove prefix added by sender chain
		voucherPrefix := transfertypes.GetDenomPrefix(packet.GetSourcePort(), packet.GetSourceChannel())

		unprefixedDenom := data.Denom[len(voucherPrefix):]

		// coin denomination used in sending from the escrow address
		denom = unprefixedDenom

		// The denomination used to send the coins is either the native denom or the hash of the path
		// if the denomination is not native.
		denomTrace := transfertypes.ParseDenomTrace(unprefixedDenom)
		if denomTrace.Path != "" {
			denom = denomTrace.IBCDenom()
		}
	} else {
		prefixedDenom := transfertypes.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel()) + data.Denom
		denom = transfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()
	}
	return denom
}
