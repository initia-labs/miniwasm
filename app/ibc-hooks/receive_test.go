package wasm_hooks_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	nfttransfertypes "github.com/initia-labs/initia/x/ibc/nft-transfer/types"
	ibchooks "github.com/initia-labs/miniwasm/app/ibc-hooks"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

func Test_OnReceivePacket(t *testing.T) {
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

	ack := input.IBCHooksMiddleware.OnRecvPacket(ctx, channeltypes.Packet{
		Data: dataBz,
	}, addr)

	require.True(t, ack.Success())
}

func Test_onReceivePacket_memo(t *testing.T) {
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
	data := transfertypes.FungibleTokenPacketData{
		Denom:    "foo",
		Amount:   "10000",
		Sender:   addr.String(),
		Receiver: contractAddrBech32,
		Memo: fmt.Sprintf(`{
			"wasm": {
				"message": {
					"contract": "%s",
					"msg": {"increase":{}}
				}
			}
		}`, contractAddrBech32),
	}

	dataBz, err := json.Marshal(&data)
	require.NoError(t, err)

	// funds foo coins to the intermediate sender
	intermediateSender, err := sdk.AccAddressFromBech32(ibchooks.DeriveIntermediateSender("channel-0", data.GetSender()))
	require.NoError(t, err)
	localDenom := ibchooks.LocalDenom(channeltypes.Packet{
		Data:               dataBz,
		DestinationPort:    "wasm",
		DestinationChannel: "channel-0",
	}, data.GetDenom())
	input.Faucet.Fund(ctx, intermediateSender, sdk.NewCoin(localDenom, math.NewInt(10000)))

	// failed to due to acl
	ack := input.IBCHooksMiddleware.OnRecvPacket(ctx, channeltypes.Packet{
		Data:               dataBz,
		DestinationPort:    "wasm",
		DestinationChannel: "channel-0",
	}, addr)
	require.False(t, ack.Success())

	// set acl
	contractAddr, err := sdk.AccAddressFromBech32(contractAddrBech32)
	require.NoError(t, err)
	require.NoError(t, input.IBCHooksKeeper.SetAllowed(ctx, contractAddr, true))

	// success
	ack = input.IBCHooksMiddleware.OnRecvPacket(ctx, channeltypes.Packet{
		Data:               dataBz,
		DestinationPort:    "wasm",
		DestinationChannel: "channel-0",
	}, addr)
	fmt.Println(string(ack.Acknowledgement()))
	require.True(t, ack.Success())

	// check the contract state
	queryRes, err := input.WasmKeeper.QuerySmart(ctx, contractAddr, []byte(`{"get":{}}`))
	require.NoError(t, err)
	require.Equal(t, "1", string(queryRes))
}

func Test_onReceiveIcs20Packet_memo_migrated(t *testing.T) {
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
	data := transfertypes.FungibleTokenPacketData{
		Denom:    "foo",
		Amount:   "10000",
		Sender:   addr.String(),
		Receiver: contractAddrBech32,
		Memo: fmt.Sprintf(`{
			"wasm": {
				"message": {
					"contract": "%s",
					"msg": {"increase":{}}
				}
			}
		}`, contractAddrBech32),
	}

	dataBz, err := json.Marshal(&data)
	require.NoError(t, err)

	pk := channeltypes.Packet{
		Data:               dataBz,
		DestinationPort:    "transfer-1",
		DestinationChannel: "channel-1",
	}

	// mint for approval test
	localDenom := ibchooks.LocalDenom(pk, data.Denom)
	// set the denom migration in OPChildKeeper
	l2Denom := "l2/" + localDenom
	input.OPChildKeeper.IBCToL2DenomMap[localDenom] = l2Denom
	intermediateSender := sdk.MustAccAddressFromBech32(ibchooks.DeriveIntermediateSender(pk.DestinationChannel, data.Sender))
	input.Faucet.Fund(ctx, intermediateSender, sdk.NewInt64Coin(l2Denom, 1000000000))

	// failed to due to acl
	ack := input.IBCHooksMiddleware.OnRecvPacket(ctx, channeltypes.Packet{
		Data:               dataBz,
		DestinationPort:    "wasm",
		DestinationChannel: "channel-0",
	}, addr)
	require.False(t, ack.Success())

	// set acl
	contractAddr, err := sdk.AccAddressFromBech32(contractAddrBech32)
	require.NoError(t, err)
	require.NoError(t, input.IBCHooksKeeper.SetAllowed(ctx, contractAddr, true))

	// success
	ack = input.IBCHooksMiddleware.OnRecvPacket(ctx, pk, addr)
	require.True(t, ack.Success())

	// check the contract state
	queryRes, err := input.WasmKeeper.QuerySmart(ctx, contractAddr, []byte(`{"get":{}}`))
	require.NoError(t, err)
	require.Equal(t, "1", string(queryRes))

	// check balance of intermediate sender
	balances := input.BankKeeper.GetAllBalances(ctx, intermediateSender)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin(l2Denom, math.NewInt(1000000000-10000))), balances)
}

func Test_OnReceivePacket_ICS721(t *testing.T) {
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

	ack := input.IBCHooksMiddleware.OnRecvPacket(ctx, channeltypes.Packet{
		Data: dataBz,
	}, addr)

	require.True(t, ack.Success())
}

func Test_onReceivePacket_memo_ICS721(t *testing.T) {
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
				"message": {
					"contract": "%s",
					"msg": {"increase":{}}
				}
			}
		}`, contractAddrBech32),
	}

	dataBz, err := json.Marshal(&data)
	require.NoError(t, err)

	// failed to due to acl
	ack := input.IBCHooksMiddleware.OnRecvPacket(ctx, channeltypes.Packet{
		Data: dataBz,
	}, addr)
	require.False(t, ack.Success())

	// set acl
	contractAddr, err := sdk.AccAddressFromBech32(contractAddrBech32)
	require.NoError(t, err)
	require.NoError(t, input.IBCHooksKeeper.SetAllowed(ctx, contractAddr, true))

	// success
	ack = input.IBCHooksMiddleware.OnRecvPacket(ctx, channeltypes.Packet{
		Data: dataBz,
	}, addr)
	require.True(t, ack.Success())

	// check the contract state
	queryRes, err := input.WasmKeeper.QuerySmart(ctx, contractAddr, []byte(`{"get":{}}`))
	require.NoError(t, err)
	require.Equal(t, "1", string(queryRes))
}
