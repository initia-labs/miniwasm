package keeper_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	wasmextensionkeeper "github.com/initia-labs/miniwasm/x/wasmextension/keeper"
	wasmextensiontypes "github.com/initia-labs/miniwasm/x/wasmextension/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

func TestMsgServer_StoreCodeAdmin(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	_, _, admin := keyPubAddr()
	_, _, addr := keyPubAddr()

	code, err := os.ReadFile("../../../app/ibc-hooks/contracts/artifacts/counter-aarch64.wasm")
	require.NoError(t, err)

	opchildParams, err := input.OPChildKeeper.GetParams(ctx)
	require.NoError(t, err)
	opchildParams.Admin = admin.String()
	require.NoError(t, input.OPChildKeeper.Params.Set(ctx, opchildParams))

	wasmMsgServer := wasmextensionkeeper.NewMsgServerImpl(&input.WasmKeeper, input.OPChildKeeper)

	// not admin
	_, err = wasmMsgServer.StoreCodeAdmin(ctx, &wasmextensiontypes.MsgStoreCodeAdmin{
		Sender:       addr.String(),
		WASMByteCode: code,
	})
	require.Error(t, err)

	// invalid code
	_, err = wasmMsgServer.StoreCodeAdmin(ctx, &wasmextensiontypes.MsgStoreCodeAdmin{
		Sender:       admin.String(),
		WASMByteCode: []byte("invalid code"),
	})
	require.Error(t, err)

	// heavy code
	longCode := make([]byte, 1024*1024*10)
	_, err = wasmMsgServer.StoreCodeAdmin(ctx, &wasmextensiontypes.MsgStoreCodeAdmin{
		Sender:       admin.String(),
		WASMByteCode: longCode,
	})
	require.Contains(t, err.Error(), wasmtypes.ErrLimit.Error())

	// valid code and admin
	storeRes, err := wasmMsgServer.StoreCodeAdmin(ctx, &wasmextensiontypes.MsgStoreCodeAdmin{
		Sender:       admin.String(),
		WASMByteCode: code,
	})
	require.NoError(t, err)
	require.NotNil(t, storeRes)
	require.NotEmpty(t, storeRes.CodeID)
	require.NotEmpty(t, storeRes.Checksum)
}
