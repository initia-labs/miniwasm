package wasm_hooks_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	nfttransfertypes "github.com/initia-labs/initia/x/ibc/nft-transfer/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

func Test_OnTimeoutPacket(t *testing.T) {
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

	err = input.IBCHooksMiddleware.OnTimeoutPacket(ctx, channeltypes.Packet{
		Data: dataBz,
	}, addr)
	require.NoError(t, err)
}

func Test_OnTimeoutPacket_memo(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	_, _, addr := keyPubAddr()

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

	// hook should not be called to due to acl
	err = input.IBCHooksMiddleware.OnTimeoutPacket(ctx, channeltypes.Packet{
		Data: dataBz,
	}, addr)
	require.NoError(t, err)

	queryRes, err := input.WasmKeeper.QuerySmart(ctx, contractAddr, []byte(`{"get":{}}`))
	require.NoError(t, err)
	require.Equal(t, "0", string(queryRes))

	// set acl
	require.NoError(t, input.IBCHooksKeeper.SetAllowed(ctx, contractAddr, true))

	// success
	err = input.IBCHooksMiddleware.OnTimeoutPacket(ctx, channeltypes.Packet{
		Data:     dataBz,
		Sequence: 99,
	}, addr)
	require.NoError(t, err)

	// check the contract state
	queryRes, err = input.WasmKeeper.QuerySmart(ctx, contractAddr, []byte(`{"get":{}}`))
	require.NoError(t, err)
	require.Equal(t, "99", string(queryRes))
}

func Test_OnTimeoutPacket_ICS721(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	_, _, addr := keyPubAddr()
	_, _, addr2 := keyPubAddr()

	data := nfttransfertypes.NonFungibleTokenPacketDataWasm{
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

	err = input.IBCHooksMiddleware.OnTimeoutPacket(ctx, channeltypes.Packet{
		Data: dataBz,
	}, addr)
	require.NoError(t, err)
}

func Test_OnTimeoutPacket_memo_ICS721(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	_, _, addr := keyPubAddr()

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

	data := nfttransfertypes.NonFungibleTokenPacketDataWasm{
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

	// success with success ack
	err = input.IBCHooksMiddleware.OnTimeoutPacket(ctx, channeltypes.Packet{
		Data: dataBz,
	}, addr)
	require.NoError(t, err)

	// check the contract state
	queryRes, err := input.WasmKeeper.QuerySmart(ctx, contractAddr, []byte(`{"get":{}}`))
	require.NoError(t, err)
	require.Equal(t, "0", string(queryRes))

	// set acl
	require.NoError(t, input.IBCHooksKeeper.SetAllowed(ctx, contractAddr, true))

	// success
	err = input.IBCHooksMiddleware.OnTimeoutPacket(ctx, channeltypes.Packet{
		Data:     dataBz,
		Sequence: 99,
	}, addr)
	require.NoError(t, err)

	// check the contract state
	queryRes, err = input.WasmKeeper.QuerySmart(ctx, contractAddr, []byte(`{"get":{}}`))
	require.NoError(t, err)
	require.Equal(t, "99", string(queryRes))
}
