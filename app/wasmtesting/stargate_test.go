package wasm_hooks_test

import (
	"os"
	"testing"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"

	slinkytypes "github.com/skip-mev/slinky/pkg/types"
)

func Test_StargateQuery(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	_, _, addr := keyPubAddr()

	code, err := os.ReadFile("./contracts/artifacts/slinky.wasm")
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
		Label:  "Slinky",
		Msg:    []byte("{}"),
		Funds:  nil,
	})
	require.NoError(t, err)

	contractAddrBech32 := instantiateRes.Address
	contractAddr, err := sdk.AccAddressFromBech32(contractAddrBech32)
	require.NoError(t, err)

	err = input.OracleKeeper.CreateCurrencyPair(ctx, slinkytypes.CurrencyPair{
		Base:  "BITCOIN",
		Quote: "USD",
	})
	require.NoError(t, err)

	res, err := input.WasmKeeper.QuerySmart(ctx, contractAddr, []byte(`{"get_all_currency_pairs": {}}`))
	require.NoError(t, err)
	require.Equal(t, "{\"currency_pairs\":[{\"Base\":\"BITCOIN\",\"Quote\":\"USD\"}]}", string(res))
}
