package keeper_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	bankkeeper "github.com/initia-labs/miniwasm/x/bank/keeper"
	tokenFactorykeeper "github.com/initia-labs/miniwasm/x/tokenfactory/keeper"
	"github.com/initia-labs/miniwasm/x/tokenfactory/types"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
)

type SendMsgTestCase struct {
	desc       string
	msg        func(denom string) *banktypes.MsgSend
	expectPass bool
}

func TestBeforeSendHook(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		wasmFile string
		sendMsgs []SendMsgTestCase
	}{
		{
			desc:     "should not allow sending 100 amount of *any* denom",
			wasmFile: "./testdata/no100.wasm",
			sendMsgs: []SendMsgTestCase{
				{
					desc: "sending 1 of factorydenom should not error",
					msg: func(factorydenom string) *banktypes.MsgSend {
						return banktypes.NewMsgSend(
							addrs[0],
							addrs[1],
							sdk.NewCoins(sdk.NewInt64Coin(factorydenom, 1)),
						)
					},
					expectPass: true,
				},
				{
					desc: "sending 1 of non-factorydenom should not error",
					msg: func(factorydenom string) *banktypes.MsgSend {
						return banktypes.NewMsgSend(
							addrs[0],
							addrs[1],
							sdk.NewCoins(sdk.NewInt64Coin("uinit", 1)),
						)
					},
					expectPass: true,
				},
				{
					desc: "sending 100 of factorydenom should error",
					msg: func(factorydenom string) *banktypes.MsgSend {
						return banktypes.NewMsgSend(
							addrs[0],
							addrs[1],
							sdk.NewCoins(sdk.NewInt64Coin(factorydenom, 100)),
						)
					},
					expectPass: false,
				},
				{
					desc: "sending 100 of non-factorydenom should work",
					msg: func(factorydenom string) *banktypes.MsgSend {
						return banktypes.NewMsgSend(
							addrs[0],
							addrs[1],
							sdk.NewCoins(sdk.NewInt64Coin("uinit", 100)),
						)
					},
					expectPass: true,
				},
				{
					desc: "having 100 coin within coins should not work",
					msg: func(factorydenom string) *banktypes.MsgSend {
						return banktypes.NewMsgSend(
							addrs[0],
							addrs[1],
							sdk.NewCoins(sdk.NewInt64Coin(factorydenom, 100), sdk.NewInt64Coin("uinit", 1)),
						)
					},
					expectPass: false,
				},
			},
		},
	} {
		t.Run(fmt.Sprintf("Case %s", tc.desc), func(t *testing.T) {
			ctx, input := createDefaultTestInput(t)

			msgServer := tokenFactorykeeper.NewMsgServerImpl(input.TokenFactoryKeeper)
			bankMsgServer := bankkeeper.NewMsgServerImpl(input.BankKeeper)

			// upload and instantiate wasm code
			wasmCode, err := os.ReadFile(tc.wasmFile)
			require.NoError(t, err, "test: %v", tc.desc)
			codeID, _, err := input.ContractKeeper.Create(ctx, addrs[0], wasmCode, nil)
			require.NoError(t, err, "test: %v", tc.desc)
			cosmwasmAddress, _, err := input.ContractKeeper.Instantiate(ctx, codeID, addrs[0], addrs[0], []byte("{}"), "", sdk.NewCoins())
			require.NoError(t, err, "test: %v", tc.desc)

			// create new denom
			res, err := msgServer.CreateDenom(ctx, types.NewMsgCreateDenom(addrs[0].String(), "bitcoin"))
			require.NoError(t, err, "test: %v", tc.desc)
			denom := res.GetNewTokenDenom()

			// mint enough coins to the creator
			_, err = msgServer.Mint(ctx, types.NewMsgMint(addrs[0].String(), sdk.NewInt64Coin(denom, 1000000000)))
			require.NoError(t, err)
			// mint some non token factory denom coins for testing
			input.Faucet.Fund(ctx, addrs[0], sdk.Coins{sdk.NewInt64Coin("uinit", 100000000000)}...)

			// set beforesend hook to the new denom
			_, err = msgServer.SetBeforeSendHook(ctx, types.NewMsgSetBeforeSendHook(addrs[0].String(), denom, cosmwasmAddress.String()))
			require.NoError(t, err, "test: %v", tc.desc)

			for _, sendTc := range tc.sendMsgs {
				_, err := bankMsgServer.Send(ctx, sendTc.msg(denom))
				if sendTc.expectPass {
					require.NoError(t, err, "test: %v", sendTc.desc)
				} else {
					require.Error(t, err, "test: %v", sendTc.desc)
				}

				// this is a check to ensure bank keeper wired in token factory keeper has hooks properly set
				// to check this, we try triggering bank hooks via token factory keeper
				for _, coin := range sendTc.msg(denom).Amount {
					_, err = msgServer.Mint(ctx, types.NewMsgMint(addrs[0].String(), sdk.NewInt64Coin(coin.Denom, coin.Amount.Int64())))
					if coin.Denom == denom && coin.Amount.Equal(sdkmath.NewInt(100)) {
						require.Error(t, err, "test: %v", sendTc.desc)
					}
				}

			}
		})
	}
}

// TestInfiniteTrackBeforeSend tests gas metering with infinite loop contract
// to properly test if we are gas metering trackBeforeSend properly.
func TestInfiniteTrackBeforeSend(t *testing.T) {
	for _, tc := range []struct {
		name               string
		wasmFile           string
		tokenToSend        sdk.Coins
		useFactoryDenom    bool
		blockBeforeSend    bool
		expectedError      bool
		useInvalidContract bool
	}{
		{
			name:            "sending tokenfactory denom from module to module with infinite contract should panic",
			wasmFile:        "./testdata/infinite_track_beforesend.wasm",
			useFactoryDenom: true,
		},
		{
			name:            "sending tokenfactory denom from module to module with infinite contract should panic",
			wasmFile:        "./testdata/infinite_track_beforesend.wasm",
			useFactoryDenom: true,
			blockBeforeSend: true,
		},
		{
			name:            "sending non-tokenfactory denom from module to module with infinite contract should not panic",
			wasmFile:        "./testdata/infinite_track_beforesend.wasm",
			tokenToSend:     sdk.NewCoins(sdk.NewInt64Coin("foo", 1000000)),
			useFactoryDenom: false,
		},
		{
			name:            "Try using no 100 ",
			wasmFile:        "./testdata/no100.wasm",
			useFactoryDenom: true,
		},
		{
			name:               "Try using invalid contract",
			wasmFile:           "./testdata/no100.wasm",
			useFactoryDenom:    true,
			useInvalidContract: true,
			expectedError:      true,
		},
	} {
		t.Run(fmt.Sprintf("Case %s", tc.name), func(t *testing.T) {
			ctx, input := createDefaultTestInput(t)

			msgServer := tokenFactorykeeper.NewMsgServerImpl(input.TokenFactoryKeeper)

			// load wasm file
			wasmCode, err := os.ReadFile(tc.wasmFile)
			require.NoError(t, err)

			// instantiate wasm code
			codeID, _, err := input.ContractKeeper.Create(ctx, addrs[0], wasmCode, nil)
			require.NoError(t, err, "test: %v", tc.name)
			cosmwasmAddress, _, err := input.ContractKeeper.Instantiate(ctx, codeID, addrs[0], addrs[0], []byte("{}"), "", sdk.NewCoins())
			require.NoError(t, err, "test: %v", tc.name)

			// create new denom
			res, err := msgServer.CreateDenom(ctx, types.NewMsgCreateDenom(addrs[0].String(), "bitcoin"))
			require.NoError(t, err, "test: %v", tc.name)
			factoryDenom := res.GetNewTokenDenom()

			var tokenToSend sdk.Coins
			if tc.useFactoryDenom {
				tokenToSend = sdk.NewCoins(sdk.NewInt64Coin(factoryDenom, 100))
			} else {
				tokenToSend = tc.tokenToSend
			}

			// send the mint module tokenToSend
			if tc.blockBeforeSend {
				input.Faucet.Fund(ctx, addrs[0], tokenToSend...)
			} else {
				input.Faucet.Fund(ctx, input.AccountKeeper.GetModuleAccount(ctx, authtypes.Minter).GetAddress(), tokenToSend...)
			}

			if tc.useInvalidContract {
				cosmwasmAddress = make(sdk.AccAddress, 32)
			}

			// set beforesend hook to the new denom
			// we register infinite loop contract here to test if we are gas metering properly
			_, err = msgServer.SetBeforeSendHook(ctx, types.NewMsgSetBeforeSendHook(addrs[0].String(), factoryDenom, cosmwasmAddress.String()))
			if tc.useInvalidContract {
				require.Error(t, err, "test: %v", tc.name)
				return
			}

			require.NoError(t, err, "test: %v", tc.name)

			if tc.blockBeforeSend {
				err = input.BankKeeper.SendCoins(ctx, addrs[0], addrs[1], tokenToSend)
				require.Error(t, err)
			} else {
				// track before send suppresses in any case, thus we expect no error
				err = input.BankKeeper.SendCoinsFromModuleToModule(ctx, authtypes.Minter, govtypes.ModuleName, tokenToSend)
				require.NoError(t, err)

				// send should happen regardless of trackBeforeSend results
				govModuleAddress := input.AccountKeeper.GetModuleAddress(govtypes.ModuleName)
				govModuleBalances := input.BankKeeper.GetAllBalances(ctx, govModuleAddress)
				require.True(t, govModuleBalances.Equal(tokenToSend))
			}

		})
	}
}
