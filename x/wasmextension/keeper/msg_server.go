package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	opchildkeeper "github.com/initia-labs/OPinit/x/opchild/keeper"

	"github.com/initia-labs/miniwasm/x/wasmextension/types"
)

var _ types.MsgServer = msgServer{}

// grpc message server implementation
type msgServer struct {
	keeper        *wasmkeeper.Keeper
	opchildKeeper *opchildkeeper.Keeper
}

// NewMsgServerImpl default constructor
func NewMsgServerImpl(k *wasmkeeper.Keeper, opchildKeeper *opchildkeeper.Keeper) types.MsgServer {
	return &msgServer{keeper: k, opchildKeeper: opchildKeeper}
}

// StoreCode stores a new wasm code on chain
func (m msgServer) StoreCodeAdmin(ctx context.Context, msg *types.MsgStoreCodeAdmin) (*types.MsgStoreCodeAdminResponse, error) {
	opchildParams, err := m.opchildKeeper.GetParams(ctx)
	if err != nil {
		return nil, err
	}
	if opchildParams.Admin != msg.Sender {
		return nil, errorsmod.Wrap(types.ErrUnauthorized, "sender is not the admin")
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}
	senderAddr, err := sdk.AccAddressFromBech32(msg.Sender)
	if err != nil {
		return nil, errorsmod.Wrap(err, "sender")
	}

	govPermissionKeeper := wasmkeeper.NewGovPermissionKeeper(m.keeper)

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	codeID, checksum, err := govPermissionKeeper.Create(sdkCtx, senderAddr, msg.WASMByteCode, msg.InstantiatePermission)
	if err != nil {
		return nil, err
	}

	return &types.MsgStoreCodeAdminResponse{
		CodeID:   codeID,
		Checksum: checksum,
	}, nil
}
