package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/initia-labs/miniwasm/x/wasmextension/types"
)

var _ types.MsgServer = msgServer{}

// grpc message server implementation
type msgServer struct {
	keeper    *wasmkeeper.Keeper
	authority string
}

// NewMsgServerImpl default constructor
func NewMsgServerImpl(k *wasmkeeper.Keeper, authority string) types.MsgServer {
	return &msgServer{keeper: k, authority: authority}
}

// StoreCode stores a new wasm code on chain
func (m msgServer) StoreCodeAdmin(ctx context.Context, msg *types.MsgStoreCodeAdmin) (*types.MsgStoreCodeAdminResponse, error) {
	if m.authority != msg.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", m.authority, msg.Authority)
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	senderAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errorsmod.Wrap(err, "invalid creator address")
	}

	govPermissionKeeper := wasmkeeper.NewGovPermissionKeeper(m.keeper)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	codeID, checksum, err := govPermissionKeeper.Create(sdkCtx, senderAddr, msg.WASMByteCode, msg.InstantiatePermission.ToWasmAccessConfig())
	if err != nil {
		return nil, err
	}

	return &types.MsgStoreCodeAdminResponse{
		CodeID:   codeID,
		Checksum: checksum,
	}, nil
}
