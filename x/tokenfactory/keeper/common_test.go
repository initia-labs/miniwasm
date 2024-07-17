package keeper_test

import (
	"encoding/binary"
	"testing"
	"time"

	appkeepers "github.com/initia-labs/miniwasm/app/keepers"

	"cosmossdk.io/core/address"
	"cosmossdk.io/log"
	"cosmossdk.io/math"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"

	"cosmossdk.io/x/tx/signing"
	"github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	testutilsims "github.com/cosmos/cosmos-sdk/testutil/sims"

	"github.com/cosmos/cosmos-sdk/types/module"
	authz "github.com/cosmos/cosmos-sdk/x/authz/module"

	initiaappparams "github.com/initia-labs/initia/app/params"

	codecaddress "github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	storetypes "cosmossdk.io/store/types"
	dbm "github.com/cosmos/cosmos-db"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	bankkeeper "github.com/initia-labs/miniwasm/x/bank/keeper"
	"github.com/initia-labs/miniwasm/x/tokenfactory"
	tokenfactorykeeper "github.com/initia-labs/miniwasm/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/initia-labs/miniwasm/x/tokenfactory/types"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
)

var ModuleBasics = module.NewBasicManager(
	auth.AppModuleBasic{},
	authz.AppModuleBasic{},
	bank.AppModuleBasic{},
	gov.AppModuleBasic{},
	wasm.AppModuleBasic{},
	tokenfactory.AppModuleBasic{},
)

var (
	valPubKeys = testutilsims.CreateTestPubKeys(5) //nolint:unused

	pubKeys = []crypto.PubKey{
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
		secp256k1.GenPrivKey().PubKey(),
	}

	addrs = []sdk.AccAddress{
		sdk.AccAddress(pubKeys[0].Address()),
		sdk.AccAddress(pubKeys[1].Address()),
		sdk.AccAddress(pubKeys[2].Address()),
		sdk.AccAddress(pubKeys[3].Address()),
		sdk.AccAddress(pubKeys[4].Address()),
	}

	//nolint:unused
	valAddrs = []sdk.ValAddress{
		sdk.ValAddress(pubKeys[0].Address()),
		sdk.ValAddress(pubKeys[1].Address()),
		sdk.ValAddress(pubKeys[2].Address()),
		sdk.ValAddress(pubKeys[3].Address()),
		sdk.ValAddress(pubKeys[4].Address()),
	}

	testDenoms = []string{
		"test1",
		"test2",
		"test3",
		"test4",
		"test5",
	}

	minitiaSupply = math.NewInt(100_000_000_000)
)

func MakeTestCodec(t testing.TB) codec.Codec {
	return MakeEncodingConfig(t).Codec
}

func MakeEncodingConfig(_ testing.TB) initiaappparams.EncodingConfig {
	interfaceRegistry, _ := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec:          codecaddress.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix()),
			ValidatorAddressCodec: codecaddress.NewBech32Codec(sdk.GetConfig().GetBech32ValidatorAddrPrefix()),
		},
	})
	appCodec := codec.NewProtoCodec(interfaceRegistry)
	legacyAmino := codec.NewLegacyAmino()
	txConfig := tx.NewTxConfig(appCodec, tx.DefaultSignModes)

	std.RegisterInterfaces(interfaceRegistry)
	std.RegisterLegacyAminoCodec(legacyAmino)

	ModuleBasics.RegisterLegacyAminoCodec(legacyAmino)
	ModuleBasics.RegisterInterfaces(interfaceRegistry)

	return initiaappparams.EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             appCodec,
		TxConfig:          txConfig,
		Amino:             legacyAmino,
	}
}

func initialTotalSupply() sdk.Coins {
	faucetBalance := sdk.NewCoins(sdk.NewCoin("uinit", minitiaSupply))
	for _, testDenom := range testDenoms {
		faucetBalance = faucetBalance.Add(sdk.NewCoin(testDenom, minitiaSupply))
	}

	return faucetBalance
}

type TestFaucet struct {
	t                testing.TB
	bankKeeper       bankkeeper.Keeper
	sender           sdk.AccAddress
	balance          sdk.Coins
	minterModuleName string
}

func NewTestFaucet(t testing.TB, ctx sdk.Context, bankKeeper bankkeeper.Keeper, minterModuleName string, minitiaSupply ...sdk.Coin) *TestFaucet {
	require.NotEmpty(t, minitiaSupply)
	r := &TestFaucet{t: t, bankKeeper: bankKeeper, minterModuleName: minterModuleName}
	_, _, addr := keyPubAddr()
	r.sender = addr
	r.Mint(ctx, addr, minitiaSupply...)
	r.balance = minitiaSupply
	return r
}

func (f *TestFaucet) Mint(parentCtx sdk.Context, addr sdk.AccAddress, amounts ...sdk.Coin) {
	amounts = sdk.Coins(amounts).Sort()
	require.NotEmpty(f.t, amounts)
	ctx := parentCtx.WithEventManager(sdk.NewEventManager()) // discard all faucet related events
	err := f.bankKeeper.MintCoins(ctx, f.minterModuleName, amounts)
	require.NoError(f.t, err)
	err = f.bankKeeper.SendCoinsFromModuleToAccount(ctx, f.minterModuleName, addr, amounts)
	require.NoError(f.t, err)
	f.balance = f.balance.Add(amounts...)
}

func (f *TestFaucet) Fund(parentCtx sdk.Context, receiver sdk.AccAddress, amounts ...sdk.Coin) {
	require.NotEmpty(f.t, amounts)
	// ensure faucet is always filled
	if !f.balance.IsAllGTE(amounts) {
		f.Mint(parentCtx, f.sender, amounts...)
	}
	ctx := parentCtx.WithEventManager(sdk.NewEventManager()) // discard all faucet related events
	err := f.bankKeeper.SendCoins(ctx, f.sender, receiver, amounts)
	require.NoError(f.t, err)
	f.balance = f.balance.Sub(amounts...)
}

func (f *TestFaucet) NewFundedAccount(ctx sdk.Context, amounts ...sdk.Coin) sdk.AccAddress {
	_, _, addr := keyPubAddr()
	f.Fund(ctx, addr, amounts...)
	return addr
}

type TestKeepers struct {
	AccountKeeper       *authkeeper.AccountKeeper
	BankKeeper          *bankkeeper.Keeper
	ContractKeeper      *wasmkeeper.PermissionedKeeper
	WasmKeeper          *wasmkeeper.Keeper
	TokenFactoryKeeper  *tokenfactorykeeper.Keeper
	CommunityPoolKeeper *appkeepers.CommunityPoolKeeper
	EncodingConfig      initiaappparams.EncodingConfig
	Faucet              *TestFaucet
	MultiStore          storetypes.CommitMultiStore
	AddressCodec        address.Codec
}

// createDefaultTestInput common settings for createTestInput
func createDefaultTestInput(t testing.TB) (sdk.Context, TestKeepers) {
	return createTestInput(t, false)
}

// createTestInput encoders can be nil to accept the defaults, or set it to override some of the message handlers (like default)
func createTestInput(t testing.TB, isCheckTx bool) (sdk.Context, TestKeepers) {
	// Load default move config
	return _createTestInput(t, isCheckTx, dbm.NewMemDB())
}

var keyCounter uint64

// we need to make this deterministic (same every test run), as encoded address size and thus gas cost,
// depends on the actual bytes (due to ugly CanonicalAddress encoding)
func keyPubAddr() (crypto.PrivKey, crypto.PubKey, sdk.AccAddress) {
	keyCounter++
	seed := make([]byte, 8)
	binary.BigEndian.PutUint64(seed, keyCounter)

	key := ed25519.GenPrivKeyFromSecret(seed)
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	return key, pub, addr
}

// encoders can be nil to accept the defaults, or set it to override some of the message handlers (like default)
func _createTestInput(
	t testing.TB,
	isCheckTx bool,
	db dbm.DB,
) (sdk.Context, TestKeepers) {
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, tokenfactorytypes.StoreKey, govtypes.StoreKey, wasmtypes.StoreKey,
	)
	ms := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	for _, v := range keys {
		ms.MountStoreWithDB(v, storetypes.StoreTypeIAVL, db)
	}
	memKeys := storetypes.NewMemoryStoreKeys()
	for _, v := range memKeys {
		ms.MountStoreWithDB(v, storetypes.StoreTypeMemory, db)
	}

	require.NoError(t, ms.LoadLatestVersion())

	ctx := sdk.NewContext(ms, tmproto.Header{
		Height: 1234567,
		Time:   time.Date(2020, time.April, 22, 12, 0, 0, 0, time.UTC),
	}, isCheckTx, log.NewNopLogger())

	encodingConfig := MakeEncodingConfig(t)
	appCodec := encodingConfig.Codec

	maccPerms := map[string][]string{ // module account permissions
		authtypes.FeeCollectorName:   nil,
		govtypes.ModuleName:          {authtypes.Burner},
		authtypes.Minter:             {authtypes.Minter, authtypes.Burner},
		tokenfactorytypes.ModuleName: {authtypes.Minter, authtypes.Burner},
	}

	ac := authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())

	accountKeeper := authkeeper.NewAccountKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[authtypes.StoreKey]), // target store
		authtypes.ProtoBaseAccount,                          // prototype
		maccPerms,
		ac,
		sdk.GetConfig().GetBech32AccountAddrPrefix(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	blockedAddrs := make(map[string]bool)
	for acc := range maccPerms {
		blockedAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	bankKeeper := bankkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[banktypes.StoreKey]),
		accountKeeper,
		blockedAddrs,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		log.NewNopLogger(),
	)
	require.NoError(t, bankKeeper.SetParams(ctx, banktypes.DefaultParams()))

	communityPoolKeeper := appkeepers.NewCommunityPoolKeeper(bankKeeper, authtypes.FeeCollectorName)
	faucet := NewTestFaucet(t, ctx, bankKeeper, authtypes.Minter, initialTotalSupply()...)
	msgRouter := baseapp.NewMsgServiceRouter()
	msgRouter.SetInterfaceRegistry(encodingConfig.InterfaceRegistry)

	keepers := TestKeepers{
		AccountKeeper:       &accountKeeper,
		BankKeeper:          &bankKeeper,
		CommunityPoolKeeper: &communityPoolKeeper,
		EncodingConfig:      encodingConfig,
		Faucet:              faucet,
		MultiStore:          ms,
		AddressCodec:        ac,
	}

	// The last arguments can contain custom message handlers, and custom query handlers,
	// if we want to allow any custom callbacks
	wasmKeeper := wasmkeeper.NewKeeper(
		appCodec,
		runtime.NewKVStoreService(keys[wasmtypes.StoreKey]),
		keepers.AccountKeeper,
		keepers.BankKeeper,
		// we do not support staking feature, so don't need to provide these keepers
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		msgRouter,
		nil,
		t.TempDir(),
		wasmtypes.DefaultWasmConfig(),
		wasmkeeper.BuiltInCapabilities(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	require.NoError(t, wasmKeeper.SetParams(ctx, wasmtypes.DefaultParams()))
	keepers.WasmKeeper = &wasmKeeper
	contractKeeper := wasmkeeper.NewDefaultPermissionKeeper(wasmKeeper)
	keepers.ContractKeeper = contractKeeper

	tokenfactoryKeeper := tokenfactorykeeper.NewKeeper(
		ac,
		appCodec,
		runtime.NewKVStoreService(keys[tokenfactorytypes.StoreKey]),
		accountKeeper,
		keepers.BankKeeper,
		keepers.CommunityPoolKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	tokenfactoryKeeper.SetContractKeeper(contractKeeper)
	keepers.TokenFactoryKeeper = &tokenfactoryKeeper
	keepers.BankKeeper.SetHooks(keepers.TokenFactoryKeeper.Hooks())

	banktypes.RegisterMsgServer(msgRouter, bankkeeper.NewMsgServerImpl(keepers.BankKeeper))
	return ctx, keepers
}
