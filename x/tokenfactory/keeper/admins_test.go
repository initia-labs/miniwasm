package keeper_test

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	tokenFactorykeeper "github.com/initia-labs/miniwasm/x/tokenfactory/keeper"
	"github.com/initia-labs/miniwasm/x/tokenfactory/types"
)

func TestAdminMsgs(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	tokenFactoryKeeper := input.TokenFactoryKeeper
	bankKeeper := input.BankKeeper
	msgServer := tokenFactorykeeper.NewMsgServerImpl(*tokenFactoryKeeper)
	queryClient := tokenFactorykeeper.Querier{Keeper: tokenFactoryKeeper}

	res, _ := msgServer.CreateDenom(ctx, types.NewMsgCreateDenom(addrs[0].String(), "bitcoin"))
	defaultDenom := res.GetNewTokenDenom()

	addr0bal := int64(0)
	addr1bal := int64(0)

	// Make sure that the admin is set correctly
	queryRes, err := queryClient.DenomAuthorityMetadata(ctx, &types.QueryDenomAuthorityMetadataRequest{
		Denom: defaultDenom,
	})
	require.NoError(t, err)
	require.Equal(t, addrs[0].String(), queryRes.AuthorityMetadata.Admin)

	// Test minting to admins own account
	_, err = msgServer.Mint(ctx, types.NewMsgMint(addrs[0].String(), sdk.NewInt64Coin(defaultDenom, 10)))
	addr0bal += 10
	require.NoError(t, err)
	require.True(t, bankKeeper.GetBalance(ctx, addrs[0], defaultDenom).Amount.Int64() == addr0bal, bankKeeper.GetBalance(ctx, addrs[0], defaultDenom))

	// Test minting to a different account
	_, err = msgServer.Mint(ctx, types.NewMsgMintTo(addrs[0].String(), sdk.NewInt64Coin(defaultDenom, 10), addrs[1].String()))
	addr1bal += 10
	require.NoError(t, err)
	require.True(t, bankKeeper.GetBalance(ctx, addrs[1], defaultDenom).Amount.Int64() == addr1bal, bankKeeper.GetBalance(ctx, addrs[1], defaultDenom))

	// Test force transferring
	_, err = msgServer.ForceTransfer(ctx, types.NewMsgForceTransfer(addrs[0].String(), sdk.NewInt64Coin(defaultDenom, 5), addrs[1].String(), addrs[0].String()))
	addr1bal -= 5
	addr0bal += 5
	require.NoError(t, err)
	require.True(t, bankKeeper.GetBalance(ctx, addrs[0], defaultDenom).Amount.Int64() == addr0bal, bankKeeper.GetBalance(ctx, addrs[0], defaultDenom))
	require.True(t, bankKeeper.GetBalance(ctx, addrs[1], defaultDenom).Amount.Int64() == addr1bal, bankKeeper.GetBalance(ctx, addrs[1], defaultDenom))

	// Test burning from own account
	_, err = msgServer.Burn(ctx, types.NewMsgBurn(addrs[0].String(), sdk.NewInt64Coin(defaultDenom, 5)))
	require.NoError(t, err)
	require.True(t, bankKeeper.GetBalance(ctx, addrs[1], defaultDenom).Amount.Int64() == addr1bal)

	// Test Change Admin
	_, err = msgServer.ChangeAdmin(ctx, types.NewMsgChangeAdmin(addrs[0].String(), defaultDenom, addrs[1].String()))
	require.NoError(t, err)
	queryRes, err = queryClient.DenomAuthorityMetadata(ctx, &types.QueryDenomAuthorityMetadataRequest{
		Denom: defaultDenom,
	})
	require.NoError(t, err)
	require.Equal(t, addrs[1].String(), queryRes.AuthorityMetadata.Admin)

	// Make sure old admin can no longer do actions
	_, err = msgServer.Burn(ctx, types.NewMsgBurn(addrs[0].String(), sdk.NewInt64Coin(defaultDenom, 5)))
	require.Error(t, err)

	// Make sure the new admin works
	_, err = msgServer.Mint(ctx, types.NewMsgMint(addrs[1].String(), sdk.NewInt64Coin(defaultDenom, 5)))
	addr1bal += 5
	require.NoError(t, err)
	require.True(t, bankKeeper.GetBalance(ctx, addrs[1], defaultDenom).Amount.Int64() == addr1bal)

	// Try setting admin to empty
	_, err = msgServer.ChangeAdmin(ctx, types.NewMsgChangeAdmin(addrs[1].String(), defaultDenom, ""))
	require.NoError(t, err)
	queryRes, err = queryClient.DenomAuthorityMetadata(ctx, &types.QueryDenomAuthorityMetadataRequest{
		Denom: defaultDenom,
	})
	require.NoError(t, err)
	require.Equal(t, "", queryRes.AuthorityMetadata.Admin)
}

// TestMintDenom ensures the following properties of the MintMessage:
// * No one can mint tokens for a denom that doesn't exist
// * Only the admin of a denom can mint tokens for it
// * The admin of a denom can mint tokens for it
func TestMintDenom(t *testing.T) {
	balances := make(map[string]int64)
	for _, acc := range addrs {
		balances[acc.String()] = 0
	}

	ctx, input := createDefaultTestInput(t)

	tokenFactoryKeeper := input.TokenFactoryKeeper
	bankKeeper := input.BankKeeper
	msgServer := tokenFactorykeeper.NewMsgServerImpl(*tokenFactoryKeeper)

	res, _ := msgServer.CreateDenom(ctx, types.NewMsgCreateDenom(addrs[0].String(), "bitcoin"))
	defaultDenom := res.GetNewTokenDenom()

	for _, tc := range []struct {
		desc       string
		mintMsg    types.MsgMint
		expectPass bool
	}{
		{
			desc: "denom does not exist",
			mintMsg: *types.NewMsgMint(
				addrs[0].String(),
				sdk.NewInt64Coin("factory/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/evmos", 10),
			),
			expectPass: false,
		},
		{
			desc: "mint is not by the admin",
			mintMsg: *types.NewMsgMintTo(
				addrs[1].String(),
				sdk.NewInt64Coin(defaultDenom, 10),
				addrs[0].String(),
			),
			expectPass: false,
		},
		{
			desc: "success case - mint to self",
			mintMsg: *types.NewMsgMint(
				addrs[0].String(),
				sdk.NewInt64Coin(defaultDenom, 10),
			),
			expectPass: true,
		},
		{
			desc: "success case - mint to another address",
			mintMsg: *types.NewMsgMintTo(
				addrs[0].String(),
				sdk.NewInt64Coin(defaultDenom, 10),
				addrs[1].String(),
			),
			expectPass: true,
		},
		{
			desc: "error: try minting non-tokenfactory denom",
			mintMsg: *types.NewMsgMintTo(
				addrs[0].String(),
				sdk.NewInt64Coin("uosmo", 10),
				addrs[1].String(),
			),
			expectPass: false,
		},
	} {
		t.Run(fmt.Sprintf("Case %s", tc.desc), func(t *testing.T) {
			_, err := msgServer.Mint(ctx, &tc.mintMsg)
			if tc.expectPass {
				require.NoError(t, err)
				balances[tc.mintMsg.MintToAddress] += tc.mintMsg.Amount.Amount.Int64()
			} else {
				require.Error(t, err)
			}

			mintToAddr, _ := sdk.AccAddressFromBech32(tc.mintMsg.MintToAddress)
			bal := bankKeeper.GetBalance(ctx, mintToAddr, defaultDenom).Amount
			require.Equal(t, bal.Int64(), balances[tc.mintMsg.MintToAddress])
		})
	}
}

func TestBurnDenom(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	tokenFactoryKeeper := input.TokenFactoryKeeper
	bankKeeper := input.BankKeeper
	accountKeeper := input.AccountKeeper
	msgServer := tokenFactorykeeper.NewMsgServerImpl(*tokenFactoryKeeper)
	res, _ := msgServer.CreateDenom(ctx, types.NewMsgCreateDenom(addrs[0].String(), "bitcoin"))
	defaultDenom := res.GetNewTokenDenom()

	// mint 1000 default token for all addrs
	balances := make(map[string]int64)
	for _, acc := range addrs {
		_, err := msgServer.Mint(ctx, types.NewMsgMintTo(addrs[0].String(), sdk.NewInt64Coin(defaultDenom, 1000), acc.String()))
		require.NoError(t, err)
		balances[acc.String()] = 1000
	}

	moduleAdress := accountKeeper.GetModuleAddress(types.ModuleName)

	for _, tc := range []struct {
		desc       string
		burnMsg    types.MsgBurn
		expectPass bool
	}{
		{
			desc: "denom does not exist",
			burnMsg: *types.NewMsgBurn(
				addrs[0].String(),
				sdk.NewInt64Coin(fmt.Sprintf("factory/%s/evmos", addrs[0].String()), 10),
			),
			expectPass: false,
		},
		{
			desc: "burn is not by the admin",
			burnMsg: *types.NewMsgBurnFrom(
				addrs[1].String(),
				sdk.NewInt64Coin(defaultDenom, 10),
				addrs[0].String(),
			),
			expectPass: false,
		},
		{
			desc: "burn more than balance",
			burnMsg: *types.NewMsgBurn(
				addrs[0].String(),
				sdk.NewInt64Coin(defaultDenom, 10000),
			),
			expectPass: false,
		},
		{
			desc: "success case - burn from self",
			burnMsg: *types.NewMsgBurn(
				addrs[0].String(),
				sdk.NewInt64Coin(defaultDenom, 10),
			),
			expectPass: true,
		},
		{
			desc: "success case - burn from another address",
			burnMsg: *types.NewMsgBurnFrom(
				addrs[0].String(),
				sdk.NewInt64Coin(defaultDenom, 10),
				addrs[1].String(),
			),
			expectPass: true,
		},
		{
			desc: "fail case - burn from module account",
			burnMsg: *types.NewMsgBurnFrom(
				addrs[0].String(),
				sdk.NewInt64Coin(defaultDenom, 10),
				moduleAdress.String(),
			),
			expectPass: false,
		},
		{
			desc: "fail case - burn non-tokenfactory denom",
			burnMsg: *types.NewMsgBurnFrom(
				addrs[0].String(),
				sdk.NewInt64Coin("uinit", 10),
				moduleAdress.String(),
			),
			expectPass: false,
		},
	} {
		t.Run(fmt.Sprintf("Case %s", tc.desc), func(t *testing.T) {
			_, err := msgServer.Burn(ctx, &tc.burnMsg)
			if tc.expectPass {
				require.NoError(t, err)
				balances[tc.burnMsg.BurnFromAddress] -= tc.burnMsg.Amount.Amount.Int64()
			} else {
				require.Error(t, err)
			}

			burnFromAddr, _ := sdk.AccAddressFromBech32(tc.burnMsg.BurnFromAddress)
			bal := bankKeeper.GetBalance(ctx, burnFromAddr, defaultDenom).Amount
			require.Equal(t, bal.Int64(), balances[tc.burnMsg.BurnFromAddress])
		})
	}
}

func TestForceTransferDenom(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	tokenFactoryKeeper := input.TokenFactoryKeeper
	bankKeeper := input.BankKeeper
	msgServer := tokenFactorykeeper.NewMsgServerImpl(*tokenFactoryKeeper)
	res, _ := msgServer.CreateDenom(ctx, types.NewMsgCreateDenom(addrs[0].String(), "bitcoin"))
	defaultDenom := res.GetNewTokenDenom()

	// mint 1000 default token for all addrs
	balances := make(map[string]int64)
	for _, acc := range addrs {
		_, err := msgServer.Mint(ctx, types.NewMsgMintTo(addrs[0].String(), sdk.NewInt64Coin(defaultDenom, 1000), acc.String()))
		require.NoError(t, err)
		balances[acc.String()] = 1000
	}

	for _, tc := range []struct {
		desc             string
		forceTransferMsg types.MsgForceTransfer
		expectPass       bool
	}{
		{
			desc: "valid force transfer",
			forceTransferMsg: *types.NewMsgForceTransfer(
				addrs[0].String(),
				sdk.NewInt64Coin(defaultDenom, 10),
				addrs[1].String(),
				addrs[2].String(),
			),
			expectPass: true,
		},
		{
			desc: "denom does not exist",
			forceTransferMsg: *types.NewMsgForceTransfer(
				addrs[0].String(),
				sdk.NewInt64Coin("factory/osmo1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44/evmos", 10),
				addrs[1].String(),
				addrs[2].String(),
			),
			expectPass: false,
		},
		{
			desc: "forceTransfer is not by the admin",
			forceTransferMsg: *types.NewMsgForceTransfer(
				addrs[1].String(),
				sdk.NewInt64Coin(defaultDenom, 10),
				addrs[1].String(),
				addrs[2].String(),
			),
			expectPass: false,
		},
		{
			desc: "forceTransfer is greater than the balance of",
			forceTransferMsg: *types.NewMsgForceTransfer(
				addrs[0].String(),
				sdk.NewInt64Coin(defaultDenom, 10000),
				addrs[1].String(),
				addrs[2].String(),
			),
			expectPass: false,
		},
	} {
		t.Run(fmt.Sprintf("Case %s", tc.desc), func(t *testing.T) {
			_, err := msgServer.ForceTransfer(ctx, &tc.forceTransferMsg)
			if tc.expectPass {
				require.NoError(t, err)

				balances[tc.forceTransferMsg.TransferFromAddress] -= tc.forceTransferMsg.Amount.Amount.Int64()
				balances[tc.forceTransferMsg.TransferToAddress] += tc.forceTransferMsg.Amount.Amount.Int64()
			} else {
				require.Error(t, err)
			}

			fromAddr, err := sdk.AccAddressFromBech32(tc.forceTransferMsg.TransferFromAddress)
			require.NoError(t, err)
			fromBal := bankKeeper.GetBalance(ctx, fromAddr, defaultDenom).Amount
			require.True(t, fromBal.Int64() == balances[tc.forceTransferMsg.TransferFromAddress])

			toAddr, err := sdk.AccAddressFromBech32(tc.forceTransferMsg.TransferToAddress)
			require.NoError(t, err)
			toBal := bankKeeper.GetBalance(ctx, toAddr, defaultDenom).Amount
			require.True(t, toBal.Int64() == balances[tc.forceTransferMsg.TransferToAddress])
		})
	}
}

func TestChangeAdminDenom(t *testing.T) {
	for _, tc := range []struct {
		desc                    string
		msgChangeAdmin          func(denom string) *types.MsgChangeAdmin
		expectedChangeAdminPass bool
		expectedAdminIndex      int
		msgMint                 func(denom string) *types.MsgMint
		expectedMintPass        bool
	}{
		{
			desc: "creator admin can't mint after setting to '' ",
			msgChangeAdmin: func(denom string) *types.MsgChangeAdmin {
				return types.NewMsgChangeAdmin(addrs[0].String(), denom, "")
			},
			expectedChangeAdminPass: true,
			expectedAdminIndex:      -1,
			msgMint: func(denom string) *types.MsgMint {
				return types.NewMsgMint(addrs[0].String(), sdk.NewInt64Coin(denom, 5))
			},
			expectedMintPass: false,
		},
		{
			desc: "non-admins can't change the existing admin",
			msgChangeAdmin: func(denom string) *types.MsgChangeAdmin {
				return types.NewMsgChangeAdmin(addrs[1].String(), denom, addrs[2].String())
			},
			expectedChangeAdminPass: false,
			expectedAdminIndex:      0,
		},
		{
			desc: "success change admin",
			msgChangeAdmin: func(denom string) *types.MsgChangeAdmin {
				return types.NewMsgChangeAdmin(addrs[0].String(), denom, addrs[1].String())
			},
			expectedAdminIndex:      1,
			expectedChangeAdminPass: true,
			msgMint: func(denom string) *types.MsgMint {
				return types.NewMsgMint(addrs[1].String(), sdk.NewInt64Coin(denom, 5))
			},
			expectedMintPass: true,
		},
	} {
		t.Run(fmt.Sprintf("Case %s", tc.desc), func(t *testing.T) {
			// setup test
			ctx, input := createDefaultTestInput(t)

			tokenFactoryKeeper := input.TokenFactoryKeeper
			msgServer := tokenFactorykeeper.NewMsgServerImpl(*tokenFactoryKeeper)
			queryClient := tokenFactorykeeper.Querier{Keeper: tokenFactoryKeeper}

			// Create a denom and mint
			res, err := msgServer.CreateDenom(ctx, types.NewMsgCreateDenom(addrs[0].String(), "bitcoin"))
			require.NoError(t, err)

			testDenom := res.GetNewTokenDenom()

			_, err = msgServer.Mint(ctx, types.NewMsgMint(addrs[0].String(), sdk.NewInt64Coin(testDenom, 10)))
			require.NoError(t, err)

			_, err = msgServer.ChangeAdmin(ctx, tc.msgChangeAdmin(testDenom))
			if tc.expectedChangeAdminPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}

			queryRes, err := queryClient.DenomAuthorityMetadata(ctx, &types.QueryDenomAuthorityMetadataRequest{
				Denom: testDenom,
			})
			require.NoError(t, err)

			// expectedAdminIndex with negative value is assumed as admin with value of ""
			const emptyStringAdminIndexFlag = -1
			if tc.expectedAdminIndex == emptyStringAdminIndexFlag {
				require.Equal(t, "", queryRes.AuthorityMetadata.Admin)
			} else {
				require.Equal(t, addrs[tc.expectedAdminIndex].String(), queryRes.AuthorityMetadata.Admin)
			}

			// we test mint to test if admin authority is performed properly after admin change.
			if tc.msgMint != nil {
				_, err := msgServer.Mint(ctx, tc.msgMint(testDenom))
				if tc.expectedMintPass {
					require.NoError(t, err)
				} else {
					require.Error(t, err)
				}
			}
		})
	}
}

func TestSetDenomMetaData(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	tokenFactoryKeeper := input.TokenFactoryKeeper
	bankKeeper := input.BankKeeper
	msgServer := tokenFactorykeeper.NewMsgServerImpl(*tokenFactoryKeeper)
	res, _ := msgServer.CreateDenom(ctx, types.NewMsgCreateDenom(addrs[0].String(), "bitcoin"))
	defaultDenom := res.GetNewTokenDenom()

	for _, tc := range []struct {
		desc                string
		msgSetDenomMetadata types.MsgSetDenomMetadata
		expectedPass        bool
	}{
		{
			desc: "successful set denom metadata",
			msgSetDenomMetadata: *types.NewMsgSetDenomMetadata(addrs[0].String(), banktypes.Metadata{
				Description: "yeehaw",
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:    defaultDenom,
						Exponent: 0,
					},
					{
						Denom:    "uinit",
						Exponent: 6,
					},
				},
				Base:    defaultDenom,
				Display: "uinit",
				Name:    "INIT",
				Symbol:  "INIT",
			}),
			expectedPass: true,
		},
		{
			desc: "non existent factory denom name",
			msgSetDenomMetadata: *types.NewMsgSetDenomMetadata(addrs[0].String(), banktypes.Metadata{
				Description: "yeehaw",
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:    fmt.Sprintf("factory/%s/litecoin", addrs[0].String()),
						Exponent: 0,
					},
					{
						Denom:    "uinit",
						Exponent: 6,
					},
				},
				Base:    fmt.Sprintf("factory/%s/litecoin", addrs[0].String()),
				Display: "uinit",
				Name:    "INIT",
				Symbol:  "INIT",
			}),
			expectedPass: false,
		},
		{
			desc: "non-factory denom",
			msgSetDenomMetadata: *types.NewMsgSetDenomMetadata(addrs[0].String(), banktypes.Metadata{
				Description: "yeehaw",
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:    "uinit",
						Exponent: 0,
					},
					{
						Denom:    "uinitt",
						Exponent: 6,
					},
				},
				Base:    "uinit",
				Display: "uinitt",
				Name:    "INIT",
				Symbol:  "INIT",
			}),
			expectedPass: false,
		},
		{
			desc: "wrong admin",
			msgSetDenomMetadata: *types.NewMsgSetDenomMetadata(addrs[1].String(), banktypes.Metadata{
				Description: "yeehaw",
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:    defaultDenom,
						Exponent: 0,
					},
					{
						Denom:    "uinit",
						Exponent: 6,
					},
				},
				Base:    defaultDenom,
				Display: "uinit",
				Name:    "INIT",
				Symbol:  "INIT",
			}),
			expectedPass: false,
		},
		{
			desc: "invalid metadata (missing display denom unit)",
			msgSetDenomMetadata: *types.NewMsgSetDenomMetadata(addrs[0].String(), banktypes.Metadata{
				Description: "yeehaw",
				DenomUnits: []*banktypes.DenomUnit{
					{
						Denom:    defaultDenom,
						Exponent: 0,
					},
				},
				Base:    defaultDenom,
				Display: "uinit",
				Name:    "INIT",
				Symbol:  "INIT",
			}),
			expectedPass: false,
		},
	} {
		t.Run(fmt.Sprintf("Case %s", tc.desc), func(t *testing.T) {
			bankKeeper := bankKeeper
			res, err := msgServer.SetDenomMetadata(ctx, &tc.msgSetDenomMetadata)
			if tc.expectedPass {
				require.NoError(t, err)
				require.NotNil(t, res)

				md, found := bankKeeper.GetDenomMetaData(ctx, defaultDenom)
				require.True(t, found)
				require.Equal(t, tc.msgSetDenomMetadata.Metadata.Name, md.Name)
			} else {
				require.Error(t, err)
			}
		})
	}
}
