package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/initia-labs/miniwasm/x/tokenfactory/types"
)

func (k Keeper) mintTo(ctx context.Context, amount sdk.Coin, mintTo string) error {
	// verify that denom is an x/tokenfactory denom
	_, _, err := types.DeconstructDenom(k.ac, amount.Denom)
	if err != nil {
		return err
	}

	err = k.bankKeeper.MintCoins(ctx, types.ModuleName, sdk.NewCoins(amount))
	if err != nil {
		return err
	}

	addr, err := k.ac.StringToBytes(mintTo)
	if err != nil {
		return err
	}

	if k.bankKeeper.BlockedAddr(addr) {
		return fmt.Errorf("failed to mint to blocked address: %s", addr)
	}

	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName,
		addr,
		sdk.NewCoins(amount))
}

func (k Keeper) burnFrom(ctx context.Context, amount sdk.Coin, burnFrom string) error {
	// verify that denom is an x/tokenfactory denom
	_, _, err := types.DeconstructDenom(k.ac, amount.Denom)
	if err != nil {
		return err
	}

	addr, err := k.ac.StringToBytes(burnFrom)
	if err != nil {
		return err
	}

	if k.bankKeeper.BlockedAddr(addr) {
		return fmt.Errorf("failed to burn from blocked address: %s", addr)
	}

	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx,
		addr,
		types.ModuleName,
		sdk.NewCoins(amount))
	if err != nil {
		return err
	}

	return k.bankKeeper.BurnCoins(ctx, types.ModuleName, sdk.NewCoins(amount))
}
