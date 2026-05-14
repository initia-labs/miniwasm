package wasm_hooks_test

import (
	"context"
	"os"
	"testing"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"

	"github.com/stretchr/testify/require"

	connecttypes "github.com/skip-mev/connect/v2/pkg/types"
)

func Test_StargateQuery(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	_, _, addr := keyPubAddr()

	code, err := os.ReadFile("./contracts/artifacts/connect.wasm")
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
		Label:  "connect",
		Msg:    []byte("{}"),
		Funds:  nil,
	})
	require.NoError(t, err)

	contractAddrBech32 := instantiateRes.Address
	contractAddr, err := sdk.AccAddressFromBech32(contractAddrBech32)
	require.NoError(t, err)

	err = input.OracleKeeper.CreateCurrencyPair(ctx, connecttypes.CurrencyPair{
		Base:  "BITCOIN",
		Quote: "USD",
	})
	require.NoError(t, err)

	res, err := input.WasmKeeper.QuerySmart(ctx, contractAddr, []byte(`{"get_all_currency_pairs": {}}`))
	require.NoError(t, err)
	require.Equal(t, "{\"currency_pairs\":[{\"Base\":\"BITCOIN\",\"Quote\":\"USD\"}]}", string(res))
}

func Test_StargateIBCTransfer(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	_, _, addr := keyPubAddr()

	var captured *transfertypes.MsgTransfer
	transfertypes.RegisterMsgServer(input.MsgRouter, &captureTransferMsgServer{
		captured: &captured,
	})

	code, err := os.ReadFile("./stargate-ibc-transfer/artifacts/stargate_ibc_transfer.wasm")
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
		Label:  "stargate-ibc-transfer",
		Msg:    []byte("{}"),
		Funds:  nil,
	})
	require.NoError(t, err)

	contractAddrBech32 := instantiateRes.Address
	contractAddr, err := sdk.AccAddressFromBech32(contractAddrBech32)
	require.NoError(t, err)

	input.Faucet.Mint(ctx, addr, sdk.NewInt64Coin("uinit", 123))

	executeMsg := []byte(`{
		"ibc_transfer": {
			"source_channel": "channel-0",
			"receiver": "cosmos1receiver",
			"denom": "uinit",
			"amount": "123",
			"timeout_timestamp": 1700000000000000000,
			"memo": "memo"
		}
	}`)

	_, err = wasmMsgServer.ExecuteContract(ctx, &wasmtypes.MsgExecuteContract{
		Sender:   addr.String(),
		Contract: contractAddrBech32,
		Msg:      executeMsg,
		Funds:    sdk.NewCoins(sdk.NewInt64Coin("uinit", 123)),
	})
	require.NoError(t, err)

	require.NotNil(t, captured)
	require.Equal(t, "transfer", captured.SourcePort)
	require.Equal(t, "channel-0", captured.SourceChannel)
	require.Equal(t, sdk.NewInt64Coin("uinit", 123), captured.Token)
	require.Equal(t, contractAddr.String(), captured.Sender)
	require.Equal(t, "cosmos1receiver", captured.Receiver)
	require.True(t, captured.TimeoutHeight.IsZero())
	require.Equal(t, uint64(1700000000000000000), captured.TimeoutTimestamp)
	require.Equal(t, "memo", captured.Memo)
}

type captureTransferMsgServer struct {
	transfertypes.UnimplementedMsgServer

	captured **transfertypes.MsgTransfer
}

func (m captureTransferMsgServer) Transfer(_ context.Context, msg *transfertypes.MsgTransfer) (*transfertypes.MsgTransferResponse, error) {
	*m.captured = msg
	return &transfertypes.MsgTransferResponse{Sequence: 1}, nil
}
