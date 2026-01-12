package wasm_hooks

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	nfttransfertypes "github.com/initia-labs/initia/x/ibc/nft-transfer/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

func Test_isIcs20Packet(t *testing.T) {
	transferMsg := transfertypes.NewFungibleTokenPacketData("denom", "1000000", "0x1", "0x2", "memo")
	bz, err := json.Marshal(transferMsg)
	require.NoError(t, err)

	ok, _transferMsg := isIcs20Packet(bz)
	require.True(t, ok)
	require.Equal(t, transferMsg, _transferMsg)

	nftTransferMsg := nfttransfertypes.NewNonFungibleTokenPacketData("class_id", "uri", "data", []string{"1", "2", "3"}, []string{"uri1", "uri2", "uri3"}, []string{"data1", "data2", "data3"}, "sender", "receiver", "memo")
	bz, err = json.Marshal(nftTransferMsg)
	require.NoError(t, err)

	ok, _ = isIcs20Packet(bz)
	require.False(t, ok)
}

func Test_isIcs721Packet(t *testing.T) {
	nftTransferMsg := nfttransfertypes.NewNonFungibleTokenPacketData("class_id", "uri", "data", []string{"1", "2", "3"}, []string{"uri1", "uri2", "uri3"}, []string{"data1", "data2", "data3"}, "sender", "receiver", "memo")

	ok, _nftTransferMsg := isIcs721Packet(nftTransferMsg.GetBytes())
	require.True(t, ok)
	require.Equal(t, nftTransferMsg, _nftTransferMsg)

	transferMsg := transfertypes.NewFungibleTokenPacketData("denom", "1000000", "0x1", "0x2", "memo")
	ok, _ = isIcs721Packet(transferMsg.GetBytes())
	require.False(t, ok)
}

func Test_parseHookData_without_callback(t *testing.T) {
	memo := `{
			"wasm" : {
				"message": {
					"sender": "init_addr",
					"contract": "contract_addr",
					"msg": {},
					"funds": [{"denom":"foo","amount":"100"}]
				}
			}
	}`
	hookData, routed, err := parseHookData(memo)
	require.True(t, routed)
	require.NoError(t, err)
	require.NotNil(t, hookData)
	require.Equal(t, &HookData{
		Message: &wasmtypes.MsgExecuteContract{
			Sender:   "init_addr",
			Contract: "contract_addr",
			Msg:      []byte("{}"),
			Funds: sdk.Coins{{
				Denom:  "foo",
				Amount: math.NewInt(100),
			}},
		},
		AsyncCallback: "",
	}, hookData)
	require.NoError(t, validateReceiver(hookData.Message, "contract_addr"))

	// invalid receiver
	require.Error(t, validateReceiver(hookData.Message, "invalid_addr"))

	hookData, routed, err = parseHookData("hihi")
	require.False(t, routed)
	require.NoError(t, err)
	require.Nil(t, hookData)
}

func Test_parseHookData_with_callback(t *testing.T) {
	memo := `{
			"wasm" : {
				"message": {
					"sender": "init_addr",
					"contract": "contract_addr",
					"msg": {},
					"funds": [{"denom":"foo","amount":"100"}]
				},
				"async_callback": "callback_addr"
			}
	}`
	hookData, routed, err := parseHookData(memo)
	require.True(t, routed)
	require.NoError(t, err)
	require.NotNil(t, hookData)
	require.Equal(t, &HookData{
		Message: &wasmtypes.MsgExecuteContract{
			Sender:   "init_addr",
			Contract: "contract_addr",
			Msg:      []byte("{}"),
			Funds: sdk.Coins{{
				Denom:  "foo",
				Amount: math.NewInt(100),
			}},
		},
		AsyncCallback: "callback_addr",
	}, hookData)
	require.NoError(t, validateReceiver(hookData.Message, "contract_addr"))
}
