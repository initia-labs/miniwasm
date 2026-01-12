package wasm_hooks_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"

	nfttransfertypes "github.com/initia-labs/initia/x/ibc/nft-transfer/types"
)

func Test_SendPacket_asyncCallback_only(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	_, _, addr := keyPubAddr()
	_, _, addr2 := keyPubAddr()

	input.MockIBCMiddleware.setSequence(42)

	data := transfertypes.FungibleTokenPacketData{
		Denom:    "foo",
		Amount:   "10000",
		Sender:   addr.String(),
		Receiver: addr2.String(),
		Memo:     fmt.Sprintf(`{"wasm":{"async_callback":"%s"},"key":"value"}`, addr.String()),
	}
	dataBz, err := json.Marshal(&data)
	require.NoError(t, err)

	seq, err := input.IBCHooksMiddleware.ICS4Middleware.SendPacket(ctx, nil, "transfer", "channel-0", clienttypes.ZeroHeight(), 0, dataBz)
	require.NoError(t, err)
	require.Equal(t, uint64(42), seq)

	var gotData transfertypes.FungibleTokenPacketData
	require.NoError(t, json.Unmarshal(input.MockIBCMiddleware.lastData, &gotData))
	require.Equal(t, `{"key":"value"}`, gotData.Memo)

	callbackBz, err := input.IBCHooksKeeper.GetAsyncCallback(ctx, "transfer", "channel-0", seq)
	require.NoError(t, err)
	expectedCallbackBz, err := json.Marshal(addr.String())
	require.NoError(t, err)
	require.Equal(t, expectedCallbackBz, callbackBz)
}

func Test_SendPacket_asyncCallback_with_message(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	_, _, addr := keyPubAddr()
	_, _, addr2 := keyPubAddr()

	input.MockIBCMiddleware.setSequence(7)

	memo := fmt.Sprintf(`{
		"wasm": {
			"message": {
				"sender": "%s",
				"contract": "%s",
				"msg": {"increase":{}}
			},
			"async_callback": "%s"
		},
		"key":"value"
	}`, addr.String(), addr2.String(), addr.String())
	data := transfertypes.FungibleTokenPacketData{
		Denom:    "foo",
		Amount:   "10000",
		Sender:   addr.String(),
		Receiver: addr2.String(),
		Memo:     memo,
	}
	dataBz, err := json.Marshal(&data)
	require.NoError(t, err)

	seq, err := input.IBCHooksMiddleware.ICS4Middleware.SendPacket(ctx, nil, "transfer", "channel-1", clienttypes.ZeroHeight(), 0, dataBz)
	require.NoError(t, err)
	require.Equal(t, uint64(7), seq)

	var gotData transfertypes.FungibleTokenPacketData
	require.NoError(t, json.Unmarshal(input.MockIBCMiddleware.lastData, &gotData))

	var memoMap map[string]any
	require.NoError(t, json.Unmarshal([]byte(gotData.Memo), &memoMap))
	wasmMap, ok := memoMap["wasm"].(map[string]any)
	require.True(t, ok)
	_, hasAsync := wasmMap["async_callback"]
	require.False(t, hasAsync)
	_, hasMessage := wasmMap["message"]
	require.True(t, hasMessage)
	keyMap, ok := memoMap["key"].(string)
	require.True(t, ok)
	require.Equal(t, "value", keyMap)

	callbackBz, err := input.IBCHooksKeeper.GetAsyncCallback(ctx, "transfer", "channel-1", seq)
	require.NoError(t, err)
	expectedCallbackBz, err := json.Marshal(addr.String())
	require.NoError(t, err)
	require.Equal(t, expectedCallbackBz, callbackBz)
}

func Test_SendPacket_asyncCallback_ics721(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	_, _, addr := keyPubAddr()

	input.MockIBCMiddleware.setSequence(11)

	data := nfttransfertypes.NonFungibleTokenPacketData{
		ClassId:   "classId",
		ClassUri:  "classUri",
		ClassData: "classData",
		TokenIds:  []string{"tokenId"},
		TokenUris: []string{"tokenUri"},
		TokenData: []string{"tokenData"},
		Sender:    addr.String(),
		Receiver:  addr.String(),
		Memo:      fmt.Sprintf(`{"wasm":{"async_callback":"%s"}}`, addr.String()),
	}
	dataBz := data.GetBytes()

	seq, err := input.IBCHooksMiddleware.ICS4Middleware.SendPacket(ctx, nil, "nft-transfer", "channel-2", clienttypes.ZeroHeight(), 0, dataBz)
	require.NoError(t, err)
	require.Equal(t, uint64(11), seq)

	gotData, err := nfttransfertypes.DecodePacketData(input.MockIBCMiddleware.lastData)
	require.NoError(t, err)
	require.Equal(t, "{}", gotData.Memo)

	callbackBz, err := input.IBCHooksKeeper.GetAsyncCallback(ctx, "nft-transfer", "channel-2", seq)
	require.NoError(t, err)
	expectedCallbackBz, err := json.Marshal(addr.String())
	require.NoError(t, err)
	require.Equal(t, expectedCallbackBz, callbackBz)
}

func Test_SendPacket_not_routed_passthrough(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	_, _, addr := keyPubAddr()
	_, _, addr2 := keyPubAddr()

	input.MockIBCMiddleware.setSequence(5)

	data := transfertypes.FungibleTokenPacketData{
		Denom:    "foo",
		Amount:   "10000",
		Sender:   addr.String(),
		Receiver: addr2.String(),
		Memo:     "not-json",
	}
	dataBz, err := json.Marshal(&data)
	require.NoError(t, err)

	seq, err := input.IBCHooksMiddleware.ICS4Middleware.SendPacket(ctx, nil, "transfer", "channel-9", clienttypes.ZeroHeight(), 0, dataBz)
	require.NoError(t, err)
	require.Equal(t, uint64(5), seq)
	var sent transfertypes.FungibleTokenPacketData
	require.NoError(t, json.Unmarshal(input.MockIBCMiddleware.lastData, &sent))
	require.Equal(t, data.Memo, sent.Memo)

	_, err = input.IBCHooksKeeper.GetAsyncCallback(ctx, "transfer", "channel-9", seq)
	require.Error(t, err)
	require.True(t, errors.Is(err, collections.ErrNotFound))
}
