package keeper_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"

	tokenFactorykeeper "github.com/initia-labs/miniwasm/x/tokenfactory/keeper"
	"github.com/initia-labs/miniwasm/x/tokenfactory/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestCreateDenom(t *testing.T) {
	var (
		primaryDenom            = "uinit"
		secondaryDenom          = "test1"
		defaultDenomCreationFee = types.Params{DenomCreationFee: sdk.NewCoins(sdk.NewCoin(primaryDenom, sdkmath.NewInt(50000000)))}
		twoDenomCreationFee     = types.Params{DenomCreationFee: sdk.NewCoins(sdk.NewCoin(primaryDenom, sdkmath.NewInt(50000000)), sdk.NewCoin(secondaryDenom, sdkmath.NewInt(50000000)))}
		nilCreationFee          = types.Params{DenomCreationFee: nil}
		largeCreationFee        = types.Params{DenomCreationFee: sdk.NewCoins(sdk.NewCoin(primaryDenom, sdkmath.NewInt(5000000000)))}
	)

	for _, tc := range []struct {
		desc             string
		denomCreationFee types.Params
		setup            func(sdk.Context, types.MsgServer)
		subdenom         string
		valid            bool
	}{
		{
			desc:             "subdenom too long",
			denomCreationFee: defaultDenomCreationFee,
			subdenom:         "assadsadsadasdasdsadsadsadsadsadsadsklkadaskkkdasdasedskhanhassyeunganassfnlksdflksafjlkasd",
			valid:            false,
		},
		{
			desc:             "subdenom and creator pair already exists",
			denomCreationFee: defaultDenomCreationFee,
			setup: func(ctx sdk.Context, msgServer types.MsgServer) {
				_, err := msgServer.CreateDenom(ctx, types.NewMsgCreateDenom(addrs[0].String(), "bitcoin"))
				require.NoError(t, err)
			},
			subdenom: "bitcoin",
			valid:    false,
		},
		{
			desc:             "success case: defaultDenomCreationFee",
			denomCreationFee: defaultDenomCreationFee,
			subdenom:         "evmos",
			valid:            true,
		},
		{
			desc:             "success case: twoDenomCreationFee",
			denomCreationFee: twoDenomCreationFee,
			subdenom:         "catcoin",
			valid:            false,
		},
		{
			desc:             "success case: nilCreationFee",
			denomCreationFee: nilCreationFee,
			subdenom:         "czcoin",
			valid:            true,
		},
		{
			desc:             "account doesn't have enough to pay for denom creation fee",
			denomCreationFee: largeCreationFee,
			subdenom:         "tooexpensive",
			valid:            false,
		},
		{
			desc:             "subdenom having invalid characters",
			denomCreationFee: defaultDenomCreationFee,
			subdenom:         "bit/***///&&&/coin",
			valid:            false,
		},
	} {
		t.Run(fmt.Sprintf("Case %s", tc.desc), func(t *testing.T) {
			ctx, input := createDefaultTestInput(t)

			tokenFactoryKeeper := input.TokenFactoryKeeper
			bankKeeper := input.BankKeeper
			// Set denom creation fee in params
			input.Faucet.Fund(ctx, addrs[0], defaultDenomCreationFee.DenomCreationFee...)

			tokenFactoryKeeper.SetParams(ctx, tc.denomCreationFee) //nolint:errcheck
			denomCreationFee := tokenFactoryKeeper.GetParams(ctx).DenomCreationFee
			require.Equal(t, tc.denomCreationFee.DenomCreationFee, denomCreationFee)

			msgServer := tokenFactorykeeper.NewMsgServerImpl(tokenFactoryKeeper)
			querier := tokenFactorykeeper.Querier{Keeper: tokenFactoryKeeper}

			if tc.setup != nil {
				tc.setup(ctx, msgServer)
			}

			// note balance, create a tokenfactory denom, then note balance again
			preCreateBalance := bankKeeper.GetAllBalances(ctx, addrs[0])
			res, err := msgServer.CreateDenom(ctx, types.NewMsgCreateDenom(addrs[0].String(), tc.subdenom))
			postCreateBalance := bankKeeper.GetAllBalances(ctx, addrs[0])
			if tc.valid {
				require.NoError(t, err)
				require.True(t, preCreateBalance.Sub(postCreateBalance...).Equal(denomCreationFee))

				// Make sure that the admin is set correctly
				queryRes, err := querier.DenomAuthorityMetadata(ctx, &types.QueryDenomAuthorityMetadataRequest{
					Denom: res.GetNewTokenDenom(),
				})

				require.NoError(t, err)
				require.Equal(t, addrs[0].String(), queryRes.AuthorityMetadata.Admin)

				// Make sure that the denom metadata is initialized correctly
				metadata, found := bankKeeper.GetDenomMetaData(ctx, res.GetNewTokenDenom())
				require.True(t, found)
				require.Equal(t, banktypes.Metadata{
					DenomUnits: []*banktypes.DenomUnit{{
						Denom:    res.GetNewTokenDenom(),
						Exponent: 0,
					}},
					Base:    res.GetNewTokenDenom(),
					Display: res.GetNewTokenDenom(),
					Name:    res.GetNewTokenDenom(),
					Symbol:  res.GetNewTokenDenom(),
				}, metadata)
			} else {
				require.Error(t, err)
				// Ensure we don't charge if we expect an error
				require.True(t, preCreateBalance.Equal(postCreateBalance))
			}
		})
	}
}

func TestGasConsume(t *testing.T) {
	// It's hard to estimate exactly how much gas will be consumed when creating a
	// denom, because besides consuming the gas specified by the params, the keeper
	// also does a bunch of other things that consume gas.
	//
	// Rather, we test whether the gas consumed is within a range. Specifically,
	// the range [gasConsume, gasConsume + offset]. If the actual gas consumption
	// falls within the range for all test cases, we consider the test passed.
	//
	// In experience, the total amount of gas consumed should consume be ~30k more
	// than the set amount.
	const offset = 50000

	for _, tc := range []struct {
		desc       string
		gasConsume uint64
	}{
		{
			desc:       "gas consume zero",
			gasConsume: 0,
		},
		{
			desc:       "gas consume 1,000,000",
			gasConsume: 1_000_000,
		},
		{
			desc:       "gas consume 10,000,000",
			gasConsume: 10_000_000,
		},
		{
			desc:       "gas consume 25,000,000",
			gasConsume: 25_000_000,
		},
		{
			desc:       "gas consume 50,000,000",
			gasConsume: 50_000_000,
		},
		{
			desc:       "gas consume 200,000,000",
			gasConsume: 200_000_000,
		},
	} {
		t.Run(fmt.Sprintf("Case %s", tc.desc), func(t *testing.T) {
			ctx, input := createDefaultTestInput(t)
			// set params with the gas consume amount

			tokenFactoryKeeper := input.TokenFactoryKeeper
			tokenFactoryKeeper.SetParams(ctx, types.NewParams(nil, tc.gasConsume)) //nolint:errcheck

			// amount of gas consumed prior to the denom creation
			gasConsumedBefore := ctx.GasMeter().GasConsumed()

			msgServer := tokenFactorykeeper.NewMsgServerImpl(tokenFactoryKeeper)
			// create a denom
			_, err := msgServer.CreateDenom(ctx, types.NewMsgCreateDenom(addrs[0].String(), "larry"))
			require.NoError(t, err)

			// amount of gas consumed after the denom creation
			gasConsumedAfter := ctx.GasMeter().GasConsumed()

			// the amount of gas consumed must be within the range
			gasConsumed := gasConsumedAfter - gasConsumedBefore
			require.Greater(t, gasConsumed, tc.gasConsume)
			require.Less(t, gasConsumed, tc.gasConsume+offset)
		})
	}
}
