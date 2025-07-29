package keeper

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

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
	} else {
		// if a contract is being set, call the contract using cache context
		// to test if the contract is an existing, valid contract.
		cacheCtx, _ := sdk.UnwrapSDKContext(ctx).CacheContext()

		cwAddr, err := sdk.AccAddressFromBech32(cosmwasmAddress)
		if err != nil {
			return err
		}

		tempMsg := types.TrackBeforeSendSudoMsg{
			TrackBeforeSend: types.TrackBeforeSendMsg{},
		}
		msgBz, err := json.Marshal(tempMsg)
		if err != nil {
			return err
		}
		_, err = k.contractKeeper.Sudo(cacheCtx, cwAddr, msgBz)

		if err != nil && strings.Contains(err.Error(), "no such contract") {
			return err
		}
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
			switch r.(type) {
			case storetypes.ErrorOutOfGas:
				k.Logger(ctx).Error("out of gas in callBeforeSendListener", "error", r)
				err = types.ErrBeforeSendHookOutOfGas
			default:
				k.Logger(ctx).Error("panic in callBeforeSendListener", "error", r)
				err = errors.New("panic in callBeforeSendListener occurred")
			}
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

			// safe guard against out of gas error
			err = k.safeSudo(ctx, cwAddr, msgBz, coin.Denom)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (k Keeper) safeSudo(ctx context.Context, cwAddr sdk.AccAddress, msgBz []byte, denom string) (err error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	gasLimit := min(sdkCtx.GasMeter().GasRemaining(), types.BeforeSendHookGasLimit)
	childCtx := sdkCtx.
		WithGasMeter(storetypes.NewGasMeter(gasLimit)).
		WithEventManager(sdk.NewEventManager())

	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case storetypes.ErrorOutOfGas:
				// propagate out of gas error
				panic(r)
			default:
				k.Logger(ctx).Error("panic in callBeforeSendListener", "error", r)
				err = errors.New("panic in callBeforeSendListener occurred")
			}
		}

		// consume gas used for calling contract to the parent ctx
		sdkCtx.GasMeter().ConsumeGas(childCtx.GasMeter().GasConsumedToLimit(), "track before send gas")
		if err == nil {
			// emit events from child context to parent context only if no error is returned
			sdkCtx.EventManager().EmitEvents(childCtx.EventManager().Events())
		}
	}()

	_, err = k.contractKeeper.Sudo(childCtx, cwAddr, msgBz)
	if err != nil {
		return errorsmod.Wrapf(err, "failed to call before send hook for denom %s", denom)
	}

	return nil
}
