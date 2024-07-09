package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	"github.com/initia-labs/miniwasm/x/tokenfactory/keeper"
	"github.com/initia-labs/miniwasm/x/tokenfactory/types"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

// TestMintDenomMsg tests TypeMsgMint message is emitted on a successful mint
func TestMintDenomMsg(t *testing.T) {
	// Create a denom
	ctx, input := createDefaultTestInput(t)

	msgServer := keeper.NewMsgServerImpl(input.TokenFactoryKeeper)
	res, _ := msgServer.CreateDenom(ctx, types.NewMsgCreateDenom(addrs[0].String(), "bitcoin"))
	defaultDenom := res.GetNewTokenDenom()

	for _, tc := range []struct {
		desc                  string
		amount                int64
		mintDenom             string
		admin                 string
		valid                 bool
		expectedMessageEvents int
	}{
		{
			desc:      "denom does not exist",
			amount:    10,
			mintDenom: fmt.Sprintf("factory/%s/evmos", addrs[0]),
			admin:     addrs[0].String(),
			valid:     false,
		},
		{
			desc:                  "success case",
			amount:                10,
			mintDenom:             defaultDenom,
			admin:                 addrs[0].String(),
			valid:                 true,
			expectedMessageEvents: 1,
		},
	} {
		t.Run(fmt.Sprintf("Case %s", tc.desc), func(t *testing.T) {
			ctx := ctx.WithEventManager(sdk.NewEventManager())
			require.Equal(t, 0, len(ctx.EventManager().Events()))
			// Test mint message
			_, err := msgServer.Mint(ctx, types.NewMsgMint(tc.admin, sdk.NewInt64Coin(tc.mintDenom, 10)))
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}

			allEvents := ctx.EventManager().Events()
			actualEvents := make([]sdk.Event, 0)
			for _, event := range allEvents {
				if event.Type == types.TypeMsgMint {
					actualEvents = append(actualEvents, event)
				}
			}
			require.Equal(t, tc.expectedMessageEvents, len(actualEvents))
		})
	}
}

// TestBurnDenomMsg tests TypeMsgBurn message is emitted on a successful burn
func TestBurnDenomMsg(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	msgServer := keeper.NewMsgServerImpl(input.TokenFactoryKeeper)
	res, _ := msgServer.CreateDenom(ctx, types.NewMsgCreateDenom(addrs[0].String(), "bitcoin"))
	defaultDenom := res.GetNewTokenDenom()

	// mint 10 default token for testAcc[0]
	_, err := msgServer.Mint(ctx, types.NewMsgMint(addrs[0].String(), sdk.NewInt64Coin(defaultDenom, 10)))
	require.NoError(t, err)

	for _, tc := range []struct {
		desc                  string
		amount                int64
		burnDenom             string
		admin                 string
		valid                 bool
		expectedMessageEvents int
	}{
		{
			desc:      "denom does not exist",
			burnDenom: fmt.Sprintf("factory/%s/evmos", addrs[0]),
			admin:     addrs[0].String(),
			valid:     false,
		},
		{
			desc:                  "success case",
			burnDenom:             defaultDenom,
			admin:                 addrs[0].String(),
			valid:                 true,
			expectedMessageEvents: 1,
		},
	} {
		t.Run(fmt.Sprintf("Case %s", tc.desc), func(t *testing.T) {
			ctx := ctx.WithEventManager(sdk.NewEventManager())
			require.Equal(t, 0, len(ctx.EventManager().Events()))
			// Test burn message
			_, err := msgServer.Burn(ctx, types.NewMsgBurn(tc.admin, sdk.NewInt64Coin(tc.burnDenom, 10)))
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			// Ensure current number and type of event is emitted
			allEvents := ctx.EventManager().Events()
			actualEvents := make([]sdk.Event, 0)
			for _, event := range allEvents {
				if event.Type == types.TypeMsgBurn {
					actualEvents = append(actualEvents, event)
				}
			}
			require.Equal(t, tc.expectedMessageEvents, len(actualEvents))
		})
	}
}

func TestForceTransferMsgFromModuleAcc(t *testing.T) {
	ctx, input := createDefaultTestInput(t)
	msgServer := keeper.NewMsgServerImpl(input.TokenFactoryKeeper)

	// Create a denom
	res, _ := msgServer.CreateDenom(ctx, types.NewMsgCreateDenom(addrs[0].String(), "bitcoin"))
	defaultDenom := res.GetNewTokenDenom()

	mintAmt := sdk.NewInt64Coin(defaultDenom, 10)

	_, err := msgServer.Mint(ctx, types.NewMsgMint(addrs[0].String(), mintAmt))
	require.NoError(t, err)

	govModAcc := input.AccountKeeper.GetModuleAccount(ctx, govtypes.ModuleName)

	err = input.BankKeeper.SendCoins(ctx, addrs[0], govModAcc.GetAddress(), sdk.NewCoins(mintAmt))
	require.NoError(t, err)

	_, err = msgServer.ForceTransfer(ctx, types.NewMsgForceTransfer(addrs[0].String(), mintAmt, govModAcc.GetAddress().String(), addrs[1].String()))
	require.ErrorContains(t, err, "failed to transfer from blocked address")
}

// TestCreateDenomMsg tests TypeMsgCreateDenom message is emitted on a successful denom creation
func TestCreateDenomMsg(t *testing.T) {
	for _, tc := range []struct {
		desc                  string
		subdenom              string
		valid                 bool
		expectedMessageEvents int
	}{
		{
			desc:     "subdenom too long",
			subdenom: "assadsadsadasdasdsadsadsadsadsadsadsklkadaskkkdasdasedskhanhassyeunganassfnlksdflksafjlkasd",
			valid:    false,
		},
		{
			desc:                  "success case: defaultDenomCreationFee",
			subdenom:              "evmos",
			valid:                 true,
			expectedMessageEvents: 1,
		},
	} {
		ctx, input := createDefaultTestInput(t)
		msgServer := keeper.NewMsgServerImpl(input.TokenFactoryKeeper)

		t.Run(fmt.Sprintf("Case %s", tc.desc), func(t *testing.T) {
			ctx := ctx.WithEventManager(sdk.NewEventManager())
			require.Equal(t, 0, len(ctx.EventManager().Events()))
			// Set denom creation fee in params
			// Test create denom message
			_, err := msgServer.CreateDenom(ctx, types.NewMsgCreateDenom(addrs[0].String(), tc.subdenom))
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			// Ensure current number and type of event is emitted
			allEvents := ctx.EventManager().Events()
			actualEvents := make([]sdk.Event, 0)
			for _, event := range allEvents {
				if event.Type == types.TypeMsgCreateDenom {
					actualEvents = append(actualEvents, event)
				}
			}
			require.Equal(t, tc.expectedMessageEvents, len(actualEvents))
		})
	}
}

// TestChangeAdminDenomMsg tests TypeMsgChangeAdmin message is emitted on a successful admin change
func TestChangeAdminDenomMsg(t *testing.T) {
	for _, tc := range []struct {
		desc                    string
		msgChangeAdmin          func(denom string) *types.MsgChangeAdmin
		expectedChangeAdminPass bool
		expectedAdminIndex      int
		msgMint                 func(denom string) *types.MsgMint
		expectedMintPass        bool
		expectedMessageEvents   int
	}{
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
			expectedMessageEvents:   1,
			msgMint: func(denom string) *types.MsgMint {
				return types.NewMsgMint(addrs[1].String(), sdk.NewInt64Coin(denom, 5))
			},
			expectedMintPass: true,
		},
	} {
		t.Run(fmt.Sprintf("Case %s", tc.desc), func(t *testing.T) {
			// setup test
			ctx, input := createDefaultTestInput(t)

			msgServer := keeper.NewMsgServerImpl(input.TokenFactoryKeeper)

			ctx = ctx.WithEventManager(sdk.NewEventManager())
			require.Equal(t, 0, len(ctx.EventManager().Events()))
			// Create a denom and mint
			res, err := msgServer.CreateDenom(ctx, types.NewMsgCreateDenom(addrs[0].String(), "bitcoin"))
			require.NoError(t, err)
			testDenom := res.GetNewTokenDenom()
			_, err = msgServer.Mint(ctx, types.NewMsgMint(addrs[0].String(), sdk.NewInt64Coin(testDenom, 10)))
			require.NoError(t, err)
			// Test change admin message
			_, err = msgServer.ChangeAdmin(ctx, tc.msgChangeAdmin(testDenom))
			if tc.expectedChangeAdminPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			// Ensure current number and type of event is emitted
			allEvents := ctx.EventManager().Events()
			actualEvents := make([]sdk.Event, 0)
			for _, event := range allEvents {
				if event.Type == types.TypeMsgChangeAdmin {
					actualEvents = append(actualEvents, event)
				}
			}
			require.Equal(t, tc.expectedMessageEvents, len(actualEvents))
		})
	}
}

// TestSetDenomMetaDataMsg tests TypeMsgSetDenomMetadata message is emitted on a successful denom metadata change
func TestSetDenomMetaDataMsg(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	msgServer := keeper.NewMsgServerImpl(input.TokenFactoryKeeper)
	res, _ := msgServer.CreateDenom(ctx, types.NewMsgCreateDenom(addrs[0].String(), "bitcoin"))
	defaultDenom := res.GetNewTokenDenom()

	for _, tc := range []struct {
		desc                  string
		msgSetDenomMetadata   types.MsgSetDenomMetadata
		expectedPass          bool
		expectedMessageEvents int
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
			expectedPass:          true,
			expectedMessageEvents: 1,
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
	} {
		t.Run(fmt.Sprintf("Case %s", tc.desc), func(t *testing.T) {
			ctx := ctx.WithEventManager(sdk.NewEventManager())
			require.Equal(t, 0, len(ctx.EventManager().Events()))
			// Test set denom metadata message
			_, err := msgServer.SetDenomMetadata(ctx, &tc.msgSetDenomMetadata)
			if tc.expectedPass {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
			// Ensure current number and type of event is emitted
			allEvents := ctx.EventManager().Events()
			actualEvents := make([]sdk.Event, 0)
			for _, event := range allEvents {
				if event.Type == types.TypeMsgSetDenomMetadata {
					actualEvents = append(actualEvents, event)
				}
			}
			require.Equal(t, tc.expectedMessageEvents, len(actualEvents))
		})
	}
}

func TestMsgUpdateParams(t *testing.T) {
	ctx, input := createDefaultTestInput(t)

	// default params
	params := types.DefaultParams()

	testCases := []struct {
		name      string
		input     *types.MsgUpdateParams
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid authority",
			input: &types.MsgUpdateParams{
				Authority: "invalid",
				Params:    params,
			},
			expErr:    true,
			expErrMsg: "invalid authority",
		},
		{
			name: "send denom creation fee param",
			input: &types.MsgUpdateParams{
				Authority: input.BankKeeper.GetAuthority(),
				Params: types.Params{
					DenomCreationFee: []sdk.Coin{{Denom: "foo", Amount: math.NewInt(-1)}},
				},
			},
			expErr:    true,
			expErrMsg: "invalid denom creation fee",
		},
		{
			name: "all good",
			input: &types.MsgUpdateParams{
				Authority: input.TokenFactoryKeeper.GetAuthority(),
				Params:    params,
			},
			expErr: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := keeper.NewMsgServerImpl(input.TokenFactoryKeeper).UpdateParams(ctx, tc.input)

			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
