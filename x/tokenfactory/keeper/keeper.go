package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"

	corestoretypes "cosmossdk.io/core/store"

	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/initia-labs/miniwasm/x/tokenfactory/types"
)

type Keeper struct {
	ac           address.Codec
	cdc          codec.Codec
	storeService corestoretypes.KVStoreService

	accountKeeper  types.AccountKeeper
	bankKeeper     types.BankKeeper
	contractKeeper types.ContractKeeper

	communityPoolKeeper types.CommunityPoolKeeper

	Schema collections.Schema
	//  key = [creator,denom], value = metadata
	CreatorDenoms  collections.KeySet[collections.Pair[string, string]]
	DenomAuthority collections.Map[string, types.DenomAuthorityMetadata]
	DenomHookAddr  collections.Map[string, string]
	Params         collections.Item[types.Params]

	authority string
}

// NewKeeper returns a new instance of the x/tokenfactory keeper
func NewKeeper(
	ac address.Codec,
	cdc codec.Codec,
	storeService corestoretypes.KVStoreService,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	communityPoolKeeper types.CommunityPoolKeeper,
	authority string,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)
	k := Keeper{
		ac:           ac,
		cdc:          cdc,
		storeService: storeService,

		accountKeeper:       accountKeeper,
		bankKeeper:          bankKeeper,
		communityPoolKeeper: communityPoolKeeper,

		CreatorDenoms:  collections.NewKeySet(sb, types.CreatorDenomsPrefix, "creatordenom", collections.PairKeyCodec(collections.StringKey, collections.StringKey)),
		DenomAuthority: collections.NewMap(sb, types.DenomAuthorityPrefix, "denomauthority", collections.StringKey, codec.CollValue[types.DenomAuthorityMetadata](cdc)),
		DenomHookAddr:  collections.NewMap(sb, types.DenomHookAddrPrefix, "denomhookaddr", collections.StringKey, collections.StringValue),

		Params: collections.NewItem(sb, types.ParamsKeyPrefix, "params", codec.CollValue[types.Params](cdc)),

		authority: authority,
	}
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return k
}

// GetAuthority returns the x/tokenfactory module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a logger for the x/tokenfactory module
func (k Keeper) Logger(ctx context.Context) log.Logger {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	return sdkCtx.Logger().With("module", "x/"+types.ModuleName)
}

// Set the wasm keeper.
func (k *Keeper) SetContractKeeper(contractKeeper types.ContractKeeper) {
	k.contractKeeper = contractKeeper
}

// CreateModuleAccount creates a module account with minting and burning capabilities
// This account isn't intended to store any coins,
// it purely mints and burns them on behalf of the admin of respective denoms,
// and sends to the relevant address.
func (k Keeper) CreateModuleAccount(ctx sdk.Context) {
	k.accountKeeper.GetModuleAccount(ctx, types.ModuleName)
}
