package keeper

import (
	"context"

	"github.com/initia-labs/miniwasm/x/tokenfactory/types"
)

// GetParams returns the total set params.
func (k Keeper) GetParams(ctx context.Context) (params types.Params) {
	params, err := k.Params.Get(ctx)
	if err != nil {
		return types.Params{}
	}

	return params
}

// SetParams sets the total set of params.
func (k Keeper) SetParams(ctx context.Context, params types.Params) error {
	return k.Params.Set(ctx, params)
}
