package wasm_hooks_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	nfttransfertypes "github.com/initia-labs/initia/x/ibc/nft-transfer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

func Test_OnAckPacket(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	_, _, addr := keyPubAddr()
	_, _, addr2 := keyPubAddr()

	data := transfertypes.FungibleTokenPacketData{
		Denom:    "foo",
		Amount:   "10000",
		Sender:   addr.String(),
		Receiver: addr2.String(),
		Memo:     "",
	}

	dataBz, err := json.Marshal(&data)
	require.NoError(t, err)

	ackBz, err := json.Marshal(channeltypes.NewResultAcknowledgement([]byte{byte(1)}))
	require.NoError(t, err)

	err = input.IBCHooksMiddleware.OnAcknowledgementPacket(ctx, channeltypes.Packet{
		Data: dataBz,
	}, ackBz, addr)
	require.NoError(t, err)
}

func Test_OnAckPacket_memo(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	_, _, addr := keyPubAddr()
	sourcePort := "transfer"
	sourceChannel := "channel-0"
	sequence := uint64(99)

	code, err := os.ReadFile("./contracts/artifacts/counter-aarch64.wasm")
	require.NoError(t, err)

	wasmMsgServer := wasmkeeper.NewMsgServerImpl(&input.WasmKeeper)
	storeRes, err := wasmMsgServer.StoreCode(ctx, &wasmtypes.MsgStoreCode{
		Sender:       addr.String(),
		WASMByteCode: code,
	})
	require.NoError(t, err)

	instantiateRes, err := wasmMsgServer.InstantiateContract(ctx, &wasmtypes.MsgInstantiateContract{
		Sender: addr.String(),
		Admin:  addr.String(),
		CodeID: storeRes.CodeID,
		Label:  "Counter",
		Msg:    []byte("{}"),
		Funds:  nil,
	})
	require.NoError(t, err)

	contractAddrBech32 := instantiateRes.Address
	contractAddr, err := sdk.AccAddressFromBech32(contractAddrBech32)
	require.NoError(t, err)

	data := transfertypes.FungibleTokenPacketData{
		Denom:    "foo",
		Amount:   "10000",
		Sender:   addr.String(),
		Receiver: contractAddrBech32,
		Memo: fmt.Sprintf(`{
			"wasm": {
				"async_callback": "%s"
			}
		}`, contractAddrBech32),
	}

	dataBz, err := json.Marshal(&data)
	require.NoError(t, err)

	successAckBz := channeltypes.NewResultAcknowledgement([]byte{byte(1)}).Acknowledgement()
	failedAckBz := channeltypes.NewErrorAcknowledgement(errors.New("failed")).Acknowledgement()
	callbackBz, err := json.Marshal(contractAddrBech32)
	require.NoError(t, err)
	require.NoError(t, input.IBCHooksKeeper.SetAsyncCallback(ctx, sourcePort, sourceChannel, sequence, callbackBz))

	// hook should not be called to due to acl
	err = input.IBCHooksMiddleware.OnAcknowledgementPacket(ctx, channeltypes.Packet{
		Data:          dataBz,
		SourcePort:    sourcePort,
		SourceChannel: sourceChannel,
		Sequence:      sequence,
	}, successAckBz, addr)
	require.NoError(t, err)

	queryRes, err := input.WasmKeeper.QuerySmart(ctx, contractAddr, []byte(`{"get":{}}`))
	require.NoError(t, err)
	require.Equal(t, "0", string(queryRes))

	// set acl
	require.NoError(t, input.IBCHooksKeeper.SetAllowed(ctx, contractAddr, true))
	require.NoError(t, input.IBCHooksKeeper.SetAsyncCallback(ctx, sourcePort, sourceChannel, sequence, callbackBz))

	// success with success ack
	err = input.IBCHooksMiddleware.OnAcknowledgementPacket(ctx, channeltypes.Packet{
		Data:          dataBz,
		SourcePort:    sourcePort,
		SourceChannel: sourceChannel,
		Sequence:      sequence,
	}, successAckBz, addr)
	require.NoError(t, err)

	// check the contract state
	queryRes, err = input.WasmKeeper.QuerySmart(ctx, contractAddr, []byte(`{"get":{}}`))
	require.NoError(t, err)
	require.Equal(t, "99", string(queryRes))

	// success with failed ack
	require.NoError(t, input.IBCHooksKeeper.SetAsyncCallback(ctx, sourcePort, sourceChannel, sequence, callbackBz))
	err = input.IBCHooksMiddleware.OnAcknowledgementPacket(ctx, channeltypes.Packet{
		Data:          dataBz,
		SourcePort:    sourcePort,
		SourceChannel: sourceChannel,
		Sequence:      sequence,
	}, failedAckBz, addr)
	require.NoError(t, err)

	// check the contract state
	queryRes, err = input.WasmKeeper.QuerySmart(ctx, contractAddr, []byte(`{"get":{}}`))
	require.NoError(t, err)
	require.Equal(t, "100", string(queryRes))
}

func Test_OnAckPacket_ICS721(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	_, _, addr := keyPubAddr()
	_, _, addr2 := keyPubAddr()

	data := nfttransfertypes.NonFungibleTokenPacketData{
		ClassId:   "classId",
		ClassUri:  "classUri",
		ClassData: "classData",
		TokenIds:  []string{"tokenId"},
		TokenUris: []string{"tokenUri"},
		TokenData: []string{"tokenData"},
		Sender:    addr.String(),
		Receiver:  addr2.String(),
		Memo:      "",
	}

	dataBz, err := json.Marshal(&data)
	require.NoError(t, err)

	ackBz, err := json.Marshal(channeltypes.NewResultAcknowledgement([]byte{byte(1)}))
	require.NoError(t, err)

	err = input.IBCHooksMiddleware.OnAcknowledgementPacket(ctx, channeltypes.Packet{
		Data: dataBz,
	}, ackBz, addr)
	require.NoError(t, err)
}

func Test_OnAckPacket_memo_ICS721(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	_, _, addr := keyPubAddr()
	sourcePort := "nft-transfer"
	sourceChannel := "channel-1"
	sequence := uint64(99)

	code, err := os.ReadFile("./contracts/artifacts/counter-aarch64.wasm")
	require.NoError(t, err)

	wasmMsgServer := wasmkeeper.NewMsgServerImpl(&input.WasmKeeper)
	storeRes, err := wasmMsgServer.StoreCode(ctx, &wasmtypes.MsgStoreCode{
		Sender:       addr.String(),
		WASMByteCode: code,
	})
	require.NoError(t, err)

	instantiateRes, err := wasmMsgServer.InstantiateContract(ctx, &wasmtypes.MsgInstantiateContract{
		Sender: addr.String(),
		Admin:  addr.String(),
		CodeID: storeRes.CodeID,
		Label:  "Counter",
		Msg:    []byte("{}"),
		Funds:  nil,
	})
	require.NoError(t, err)

	contractAddrBech32 := instantiateRes.Address
	contractAddr, err := sdk.AccAddressFromBech32(contractAddrBech32)
	require.NoError(t, err)

	data := nfttransfertypes.NonFungibleTokenPacketData{
		ClassId:   "classId",
		ClassUri:  "classUri",
		ClassData: "classData",
		TokenIds:  []string{"tokenId"},
		TokenUris: []string{"tokenUri"},
		TokenData: []string{"tokenData"},
		Sender:    addr.String(),
		Receiver:  contractAddrBech32,
		Memo: fmt.Sprintf(`{
			"wasm": {
				"async_callback": "%s"
			}
		}`, contractAddrBech32),
	}

	dataBz, err := json.Marshal(&data)
	require.NoError(t, err)

	successAckBz := channeltypes.NewResultAcknowledgement([]byte{byte(1)}).Acknowledgement()
	failedAckBz := channeltypes.NewErrorAcknowledgement(errors.New("failed")).Acknowledgement()
	callbackBz, err := json.Marshal(contractAddrBech32)
	require.NoError(t, err)
	require.NoError(t, input.IBCHooksKeeper.SetAsyncCallback(ctx, sourcePort, sourceChannel, sequence, callbackBz))

	// success with success ack
	err = input.IBCHooksMiddleware.OnAcknowledgementPacket(ctx, channeltypes.Packet{
		Data:          dataBz,
		SourcePort:    sourcePort,
		SourceChannel: sourceChannel,
		Sequence:      sequence,
	}, successAckBz, addr)
	require.NoError(t, err)

	// check the contract state
	queryRes, err := input.WasmKeeper.QuerySmart(ctx, contractAddr, []byte(`{"get":{}}`))
	require.NoError(t, err)
	require.Equal(t, "0", string(queryRes))

	// set acl
	require.NoError(t, input.IBCHooksKeeper.SetAllowed(ctx, contractAddr, true))
	require.NoError(t, input.IBCHooksKeeper.SetAsyncCallback(ctx, sourcePort, sourceChannel, sequence, callbackBz))

	// success with success ack
	err = input.IBCHooksMiddleware.OnAcknowledgementPacket(ctx, channeltypes.Packet{
		Data:          dataBz,
		SourcePort:    sourcePort,
		SourceChannel: sourceChannel,
		Sequence:      sequence,
	}, successAckBz, addr)
	require.NoError(t, err)

	// check the contract state
	queryRes, err = input.WasmKeeper.QuerySmart(ctx, contractAddr, []byte(`{"get":{}}`))
	require.NoError(t, err)
	require.Equal(t, "99", string(queryRes))

	// success with failed ack
	require.NoError(t, input.IBCHooksKeeper.SetAsyncCallback(ctx, sourcePort, sourceChannel, sequence, callbackBz))
	err = input.IBCHooksMiddleware.OnAcknowledgementPacket(ctx, channeltypes.Packet{
		Data:          dataBz,
		SourcePort:    sourcePort,
		SourceChannel: sourceChannel,
		Sequence:      sequence,
	}, failedAckBz, addr)
	require.NoError(t, err)

	// check the contract state
	queryRes, err = input.WasmKeeper.QuerySmart(ctx, contractAddr, []byte(`{"get":{}}`))
	require.NoError(t, err)
	require.Equal(t, "100", string(queryRes))
}
