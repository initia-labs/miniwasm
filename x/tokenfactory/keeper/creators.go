package keeper

import (
	"context"

	"cosmossdk.io/collections"
)

func (k Keeper) addDenomFromCreator(ctx context.Context, creator, denom string) error {
	return k.CreatorDenoms.Set(ctx, collections.Join(creator, denom))
}

func (k Keeper) getDenomsFromCreator(ctx context.Context, creator string) []string {
	denoms := []string{}
	err := k.CreatorDenoms.Walk(ctx, collections.NewPrefixedPairRange[string, string](creator), func(key collections.Pair[string, string]) (stop bool, err error) {
		denoms = append(denoms, key.K2())
		return false, nil
	})
	if err != nil {
		panic(err)
	}
	return denoms
}
