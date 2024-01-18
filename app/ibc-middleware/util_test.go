package ibc_middleware

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func Test_validateAndParseMemo(t *testing.T) {
	memo := `
	{
		"wasm" : {
			"sender": "init_addr",
			"contract": "contract_addr",
			"msg": {},
			"funds": [{"denom":"foo","amount":"100"}]
		}
	}`
	isWasmRouted, msg, err := validateAndParseMemo(memo, "contract_addr")
	require.True(t, isWasmRouted)
	require.NoError(t, err)
	require.Equal(t, wasmtypes.MsgExecuteContract{
		Sender:   "init_addr",
		Contract: "contract_addr",
		Msg:      []byte("{}"),
		Funds: sdk.Coins{{
			Denom:  "foo",
			Amount: math.NewInt(100),
		}},
	}, msg)

	// invalid receiver
	isWasmRouted, _, err = validateAndParseMemo(memo, "invalid_addr")
	require.True(t, isWasmRouted)
	require.Error(t, err)

	isWasmRouted, _, err = validateAndParseMemo("hihi", "invalid_addr")
	require.False(t, isWasmRouted)
	require.NoError(t, err)
}
