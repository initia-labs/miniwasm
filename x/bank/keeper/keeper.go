package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Keeper struct {
	bankkeeper.BaseKeeper

	ak    types.AccountKeeper
	hooks BankHooks
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	ak types.AccountKeeper,
	blockedAddrs map[string]bool,
	authority string,
	logger log.Logger,
) Keeper {
	if _, err := ak.AddressCodec().StringToBytes(authority); err != nil {
		panic(fmt.Errorf("invalid bank authority address: %w", err))
	}

	return Keeper{
		BaseKeeper: bankkeeper.NewBaseKeeper(cdc, storeService, ak, blockedAddrs, authority, logger),
		ak:         ak,
	}
}

func (k Keeper) DelegateCoins(ctx context.Context, delegatorAddr, moduleAccAddr sdk.AccAddress, amt sdk.Coins) error {
	moduleAcc := k.ak.GetAccount(ctx, moduleAccAddr)
	if moduleAcc == nil {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleAccAddr)
	}

	if !amt.IsValid() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, amt.String())
	}

	err := k.BlockBeforeSend(ctx, delegatorAddr, moduleAccAddr, amt)
	if err != nil {
		return err
	}
	// call the TrackBeforeSend hooks and the BlockBeforeSend hooks
	k.TrackBeforeSend(ctx, delegatorAddr, moduleAccAddr, amt)

	return k.BaseKeeper.DelegateCoins(ctx, delegatorAddr, moduleAccAddr, amt)
}

func (k Keeper) UndelegateCoins(ctx context.Context, moduleAccAddr, delegatorAddr sdk.AccAddress, amt sdk.Coins) error {
	moduleAcc := k.ak.GetAccount(ctx, moduleAccAddr)
	if moduleAcc == nil {
		return errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleAccAddr)
	}

	if !amt.IsValid() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, amt.String())
	}

	// call the TrackBeforeSend hooks and the BlockBeforeSend hooks
	err := k.BlockBeforeSend(ctx, moduleAccAddr, delegatorAddr, amt)
	if err != nil {
		return err
	}

	k.TrackBeforeSend(ctx, moduleAccAddr, delegatorAddr, amt)

	return k.BaseKeeper.UndelegateCoins(ctx, moduleAccAddr, moduleAccAddr, amt)
}

// SendCoinsFromModuleToModule transfers coins from a ModuleAccount to another.
// It will panic if either module account does not exist.
// SendCoinsFromModuleToModule is the only send method that does not call both BlockBeforeSend and TrackBeforeSend hook.
// It only calls the TrackBeforeSend hook.
func (k Keeper) SendCoinsFromModuleToModule(
	ctx context.Context, senderModule, recipientModule string, amt sdk.Coins,
) error {
	senderAddr := k.ak.GetModuleAddress(senderModule)
	if senderAddr == nil {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", senderModule))
	}

	recipientAcc := k.ak.GetModuleAccount(ctx, recipientModule)
	if recipientAcc == nil {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", recipientModule))
	}

	// return k.SendCoins(ctx, senderAddr, recipientAcc.GetAddress(), amt)
	return k.SendCoinsWithoutBlockHook(ctx, senderAddr, recipientAcc.GetAddress(), amt)
}

func (k Keeper) SendCoinsFromModuleToAccount(
	ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins,
) error {
	senderAddr := k.ak.GetModuleAddress(senderModule)
	if senderAddr == nil {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", senderModule))
	}

	if k.BlockedAddr(recipientAddr) {
		return errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "%s is not allowed to receive funds", recipientAddr)
	}

	return k.SendCoins(ctx, senderAddr, recipientAddr, amt)
}

// SendCoinsFromAccountToModule transfers coins from an AccAddress to a ModuleAccount.
// It will panic if the module account does not exist.
func (k Keeper) SendCoinsFromAccountToModule(
	ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins,
) error {
	recipientAcc := k.ak.GetModuleAccount(ctx, recipientModule)
	if recipientAcc == nil {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", recipientModule))
	}

	return k.SendCoins(ctx, senderAddr, recipientAcc.GetAddress(), amt)
}

// DelegateCoinsFromAccountToModule delegates coins and transfers them from a
// delegator account to a module account. It will panic if the module account
// does not exist or is unauthorized.
func (k Keeper) DelegateCoinsFromAccountToModule(
	ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins,
) error {
	recipientAcc := k.ak.GetModuleAccount(ctx, recipientModule)
	if recipientAcc == nil {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", recipientModule))
	}

	if !recipientAcc.HasPermission(authtypes.Staking) {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to receive delegated coins", recipientModule))
	}

	return k.DelegateCoins(ctx, senderAddr, recipientAcc.GetAddress(), amt)
}

// UndelegateCoinsFromModuleToAccount undelegates the unbonding coins and transfers
// them from a module account to the delegator account. It will panic if the
// module account does not exist or is unauthorized.
func (k Keeper) UndelegateCoinsFromModuleToAccount(
	ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins,
) error {
	acc := k.ak.GetModuleAccount(ctx, senderModule)
	if acc == nil {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", senderModule))
	}

	if !acc.HasPermission(authtypes.Staking) {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnauthorized, "module account %s does not have permissions to undelegate coins", senderModule))
	}

	return k.UndelegateCoins(ctx, acc.GetAddress(), recipientAddr, amt)
}

type BankHooks interface {
	TrackBeforeSend(ctx context.Context, from, to sdk.AccAddress, amount sdk.Coins)       // Must be before any send is executed
	BlockBeforeSend(ctx context.Context, from, to sdk.AccAddress, amount sdk.Coins) error // Must be before any send is executed
}

// TrackBeforeSend executes the TrackBeforeSend hook if registered.
func (k Keeper) TrackBeforeSend(ctx context.Context, from, to sdk.AccAddress, amount sdk.Coins) {
	if k.hooks != nil {
		k.hooks.TrackBeforeSend(ctx, from, to, amount)
	}
}

// BlockBeforeSend executes the BlockBeforeSend hook if registered.
func (k Keeper) BlockBeforeSend(ctx context.Context, from, to sdk.AccAddress, amount sdk.Coins) error {
	if k.hooks != nil {
		return k.hooks.BlockBeforeSend(ctx, from, to, amount)
	}
	return nil
}

func (k *Keeper) SetHooks(bh BankHooks) *Keeper {
	if k.hooks != nil {
		panic("cannot set bank hooks twice")
	}
	k.hooks = bh
	return k
}

// SendCoinsWithoutBlockHook calls sendCoins without calling the `BlockBeforeSend` hook.
func (k Keeper) SendCoinsWithoutBlockHook(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error {
	// call the TrackBeforeSend hooks
	k.TrackBeforeSend(ctx, fromAddr, toAddr, amt)
	return k.BaseSendKeeper.SendCoins(ctx, fromAddr, toAddr, amt)
}

func (k Keeper) SendCoins(ctx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error {
	// BlockBeforeSend hook should always be called before the TrackBeforeSend hook.
	err := k.BlockBeforeSend(ctx, fromAddr, toAddr, amt)
	if err != nil {
		return err
	}
	// call the TrackBeforeSend hooks
	k.TrackBeforeSend(ctx, fromAddr, toAddr, amt)
	return k.BaseSendKeeper.SendCoins(ctx, fromAddr, toAddr, amt)
}
