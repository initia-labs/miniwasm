package keeper

import (
	"context"
	"encoding/json"

	storetypes "cosmossdk.io/store/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	wasmvmtypes "github.com/CosmWasm/wasmvm/v2/types"
	"github.com/initia-labs/miniwasm/x/tokenfactory/types"

	errorsmod "cosmossdk.io/errors"
)

func (k Keeper) setBeforeSendHook(ctx context.Context, denom string, cosmwasmAddress string) error {
	// verify that denom is an x/tokenfactory denom
	_, _, err := types.DeconstructDenom(k.ac, denom)
	if err != nil {
		return err
	}

	// delete the store for denom prefix store when cosmwasm address is nil
	if cosmwasmAddress == "" {
		return k.DenomHookAddr.Remove(ctx, denom)
	}

	_, err = k.ac.StringToBytes(cosmwasmAddress)
	if err != nil {
		return err
	}

	return k.DenomHookAddr.Set(ctx, denom, cosmwasmAddress)
}

func (k Keeper) GetBeforeSendHook(ctx context.Context, denom string) string {
	wasmAddr, err := k.DenomHookAddr.Get(ctx, denom)
	if err != nil {
		return ""
	}

	return wasmAddr
}

// Hooks wrapper struct for bank keeper
type Hooks struct {
	k Keeper
}

var _ types.BankHooks = Hooks{}

// Return the wrapper struct
func (k Keeper) Hooks() Hooks {
	return Hooks{k}
}

// TrackBeforeSend calls the before send listener contract suppresses any errors
func (h Hooks) TrackBeforeSend(ctx context.Context, from, to sdk.AccAddress, amount sdk.Coins) {
	_ = h.k.callBeforeSendListener(ctx, from, to, amount, false)
}

// TrackBeforeSend calls the before send listener contract returns any errors
func (h Hooks) BlockBeforeSend(ctx context.Context, from, to sdk.AccAddress, amount sdk.Coins) error {
	return h.k.callBeforeSendListener(ctx, from, to, amount, true)
}

// callBeforeSendListener iterates over each coin and sends corresponding sudo msg to the contract address stored in state.
// If blockBeforeSend is true, sudoMsg wraps BlockBeforeSendMsg, otherwise sudoMsg wraps TrackBeforeSendMsg.
// Note that we gas meter trackBeforeSend to prevent infinite contract calls.
// CONTRACT: this should not be called in beginBlock or endBlock since out of gas will cause this method to panic.
func (k Keeper) callBeforeSendListener(ctx context.Context, from, to sdk.AccAddress, amount sdk.Coins, blockBeforeSend bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errorsmod.Wrapf(types.ErrBeforeSendHookOutOfGas, "%v", r)
		}
	}()

	fromAddr, err := k.ac.BytesToString(from)
	if err != nil {
		return err
	}

	toAddr, err := k.ac.BytesToString(to)
	if err != nil {
		return err
	}

	for _, coin := range amount {
		cosmwasmAddress := k.GetBeforeSendHook(ctx, coin.Denom)
		if cosmwasmAddress != "" {
			cwAddr, err := k.ac.StringToBytes(cosmwasmAddress)
			if err != nil {
				return err
			}

			var msgBz []byte

			// get msgBz, either BlockBeforeSend or TrackBeforeSend
			// Note that for trackBeforeSend, we need to gas meter computations to prevent infinite loop
			// specifically because module to module sends are not gas metered.
			// We don't need to do this for blockBeforeSend since blockBeforeSend is not called during module to module sends.
			if blockBeforeSend {
				msg := types.BlockBeforeSendSudoMsg{
					BlockBeforeSend: types.BlockBeforeSendMsg{
						From: fromAddr,
						To:   toAddr,
						Amount: wasmvmtypes.Coin{
							Denom:  coin.GetDenom(),
							Amount: coin.Amount.String(),
						},
					},
				}
				msgBz, err = json.Marshal(msg)
			} else {
				msg := types.TrackBeforeSendSudoMsg{
					TrackBeforeSend: types.TrackBeforeSendMsg{
						From: fromAddr,
						To:   toAddr,
						Amount: wasmvmtypes.Coin{
							Denom:  coin.GetDenom(),
							Amount: coin.Amount.String(),
						},
					},
				}
				msgBz, err = json.Marshal(msg)
			}
			if err != nil {
				return err
			}
			em := sdk.NewEventManager()

			sdkCtx := sdk.UnwrapSDKContext(ctx)
			childCtx := sdkCtx.WithGasMeter(storetypes.NewGasMeter(types.BeforeSendHookGasLimit))
			_, err = k.contractKeeper.Sudo(childCtx.WithEventManager(em), cwAddr, msgBz)
			if err != nil {
				return errorsmod.Wrapf(err, "failed to call before send hook for denom %s", coin.Denom)
			}

			// consume gas used for calling contract to the parent ctx
			sdkCtx.GasMeter().ConsumeGas(childCtx.GasMeter().GasConsumed(), "track before send gas")
		}
	}
	return nil
}
