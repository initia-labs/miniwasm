package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	"github.com/initia-labs/miniwasm/x/tokenfactory/types"
)

func TestGenesis(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	tokenFactoryKeeper := input.TokenFactoryKeeper
	bankKeeper := input.BankKeeper
	accountKeeper := input.AccountKeeper

	creator, err := input.AddressCodec.BytesToString(addrs[0])
	require.NoError(t, err, "address encoding")

	another, err := input.AddressCodec.BytesToString(addrs[1])
	require.NoError(t, err, "address encoding")

	genesisState := types.GenesisState{
		FactoryDenoms: []types.GenesisDenom{
			{
				Denom: fmt.Sprintf("factory/%s/bitcoin", creator),
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: creator,
				},
			},
			{
				Denom: fmt.Sprintf("factory/%s/diff-admin", creator),
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: another,
				},
			},
			{
				Denom: fmt.Sprintf("factory/%s/litecoin", creator),
				AuthorityMetadata: types.DenomAuthorityMetadata{
					Admin: creator,
				},
			},
		},
	}

	// Test both with bank denom metadata set, and not set.
	for i, denom := range genesisState.FactoryDenoms {
		// hacky, sets bank metadata to exist if i != 0, to cover both cases.
		if i != 0 {
			bankKeeper.SetDenomMetaData(ctx, banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{{
					Denom:    denom.GetDenom(),
					Exponent: 0,
				}},
				Base:    denom.GetDenom(),
				Display: denom.GetDenom(),
				Name:    denom.GetDenom(),
				Symbol:  denom.GetDenom(),
			})
		}
	}

	// check before initGenesis that the module account is nil
	tokenfactoryModuleAccount := accountKeeper.GetAccount(ctx, accountKeeper.GetModuleAddress(types.ModuleName))
	require.Nil(t, tokenfactoryModuleAccount)

	tokenFactoryKeeper.SetParams(ctx, types.Params{DenomCreationFee: sdk.Coins{sdk.NewInt64Coin("uinit", 100)}}) //nolint:errcheck
	tokenFactoryKeeper.InitGenesis(ctx, genesisState)

	// check that the module account is now initialized
	tokenfactoryModuleAccount = accountKeeper.GetAccount(ctx, accountKeeper.GetModuleAddress(types.ModuleName))
	require.NotNil(t, tokenfactoryModuleAccount)

	exportedGenesis := tokenFactoryKeeper.ExportGenesis(ctx)
	require.NotNil(t, exportedGenesis)
	require.Equal(t, genesisState, *exportedGenesis)

	// verify that the exported bank genesis is valid
	bankKeeper.SetParams(ctx, banktypes.DefaultParams()) //nolint:errcheck
	exportedBankGenesis := bankKeeper.ExportGenesis(ctx)
	require.NoError(t, exportedBankGenesis.Validate())

	bankKeeper.InitGenesis(ctx, exportedBankGenesis)
	for i, denom := range genesisState.FactoryDenoms {
		// hacky, check whether bank metadata is not replaced if i != 0, to cover both cases.
		if i != 0 {
			metadata, found := bankKeeper.GetDenomMetaData(ctx, denom.GetDenom())
			require.True(t, found)
			require.EqualValues(t, metadata, banktypes.Metadata{
				DenomUnits: []*banktypes.DenomUnit{{
					Denom:    denom.GetDenom(),
					Exponent: 0,
				}},
				Base:    denom.GetDenom(),
				Display: denom.GetDenom(),
				Name:    denom.GetDenom(),
				Symbol:  denom.GetDenom(),
			})
		}
	}
}
