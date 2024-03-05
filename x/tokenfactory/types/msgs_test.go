package types_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/stretchr/testify/require"

	"github.com/initia-labs/miniwasm/x/tokenfactory/types"

	"github.com/cometbft/cometbft/crypto/ed25519"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"
	proto "github.com/cosmos/gogoproto/proto"
	initiaappparams "github.com/initia-labs/initia/app/params"
)

func testMessageAuthzSerialization(t *testing.T, msg sdk.Msg) {
	someDate := time.Date(1, 1, 1, 1, 1, 1, 1, time.UTC)
	const (
		mockGranter string = "cosmos1abc"
		mockGrantee string = "cosmos1xyz"
	)

	var (
		mockMsgGrant  authz.MsgGrant
		mockMsgRevoke authz.MsgRevoke
		mockMsgExec   authz.MsgExec
	)

	encodingConfig := initiaappparams.MakeEncodingConfig()
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	legacyAmino := encodingConfig.Amino
	authz.RegisterLegacyAminoCodec(legacyAmino)
	authz.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	types.RegisterLegacyAminoCodec(legacyAmino)
	types.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	// mutates mockMsg
	testSerDeser := func(msg proto.Message, mockMsg proto.Message) {
		msgGrantBytes := json.RawMessage(sdk.MustSortJSON(legacyAmino.MustMarshalJSON(msg)))
		err := legacyAmino.UnmarshalJSON(msgGrantBytes, mockMsg)
		require.NoError(t, err)
	}

	// Authz: Grant Msg
	typeURL := sdk.MsgTypeURL(msg)
	expiryTime := someDate.Add(time.Hour)
	grant, err := authz.NewGrant(someDate, authz.NewGenericAuthorization(typeURL), &expiryTime)
	require.NoError(t, err)

	msgGrant := authz.MsgGrant{Granter: mockGranter, Grantee: mockGrantee, Grant: grant}
	testSerDeser(&msgGrant, &mockMsgGrant)

	// Authz: Revoke Msg
	msgRevoke := authz.MsgRevoke{Granter: mockGranter, Grantee: mockGrantee, MsgTypeUrl: typeURL}
	testSerDeser(&msgRevoke, &mockMsgRevoke)

	// Authz: Exec Msg

	msgAny, err := codectypes.NewAnyWithValue(msg)
	require.NoError(t, err)

	msgExec := authz.MsgExec{Grantee: mockGrantee, Msgs: []*codectypes.Any{msgAny}}
	testSerDeser(&msgExec, &mockMsgExec)
	require.Equal(t, msgExec.Msgs[0].Value, mockMsgExec.Msgs[0].Value)
}

// // Test authz serialize and de-serializes for tokenfactory msg.
func TestAuthzMsg(t *testing.T) {
	pk1 := ed25519.GenPrivKey().PubKey()

	addr1 := sdk.AccAddress(pk1.Address()).String()
	coin := sdk.NewCoin("denom", sdkmath.NewInt(1))

	testCases := []struct {
		name string
		msg  sdk.Msg
	}{
		{
			name: "MsgCreateDenom",
			msg: &types.MsgCreateDenom{
				Sender:   addr1,
				Subdenom: "valoper1xyz",
			},
		},
		{
			name: "MsgBurn",
			msg: &types.MsgBurn{
				Sender: addr1,
				Amount: coin,
			},
		},
		{
			name: "MsgMint",
			msg: &types.MsgMint{
				Sender: addr1,
				Amount: coin,
			},
		},
		{
			name: "MsgChangeAdmin",
			msg: &types.MsgChangeAdmin{
				Sender:   addr1,
				Denom:    "denom",
				NewAdmin: "init1q8tq5qhrhw6t970egemuuwywhlhpnmdmts6xnu",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testMessageAuthzSerialization(t, tc.msg)
		})
	}
}

// TestMsgCreateDenom tests if valid/invalid create denom messages are properly validated/invalidated
func TestMsgCreateDenom(t *testing.T) {
	// generate a private/public key pair and get the respective address
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address())

	// make a proper createDenom message
	createMsg := func(after func(msg types.MsgCreateDenom) types.MsgCreateDenom) types.MsgCreateDenom {
		properMsg := *types.NewMsgCreateDenom(
			addr1.String(),
			"bitcoin",
		)

		return after(properMsg)
	}

	// validate createDenom message was created as intended
	msg := createMsg(func(msg types.MsgCreateDenom) types.MsgCreateDenom {
		return msg
	})

	codec := initiaappparams.MakeEncodingConfig().Codec
	ac := authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())

	signers, _, err := codec.GetMsgV1Signers(&msg)
	require.NoError(t, err, "can get signers from msg")

	signer, err := ac.BytesToString(signers[0])
	require.NoError(t, err, "can parse signer to string")
	require.Equal(t, len(signers), 1)
	require.Equal(t, signer, addr1.String())

	tests := []struct {
		name       string
		msg        types.MsgCreateDenom
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg types.MsgCreateDenom) types.MsgCreateDenom {
				return msg
			}),
			expectPass: true,
		},
		{
			name: "empty sender",
			msg: createMsg(func(msg types.MsgCreateDenom) types.MsgCreateDenom {
				msg.Sender = ""
				return msg
			}),
			expectPass: false,
		},
		{
			name: "invalid subdenom",
			msg: createMsg(func(msg types.MsgCreateDenom) types.MsgCreateDenom {
				msg.Subdenom = "thissubdenomismuchtoolongasdkfjaasdfdsafsdlkfnmlksadmflksmdlfmlsakmfdsafasdfasdf"
				return msg
			}),
			expectPass: false,
		},
	}

	for _, test := range tests {
		if test.expectPass {
			require.NoError(t, test.msg.Validate(ac), "test: %v", test.name)
		} else {
			require.Error(t, test.msg.Validate(ac), "test: %v", test.name)
		}
	}
}

// TestMsgMint tests if valid/invalid create denom messages are properly validated/invalidated
func TestMsgMint(t *testing.T) {
	// generate a private/public key pair and get the respective address
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address())

	// make a proper mint message
	createMsg := func(after func(msg types.MsgMint) types.MsgMint) types.MsgMint {
		properMsg := *types.NewMsgMint(
			addr1.String(),
			sdk.NewCoin("bitcoin", sdkmath.NewInt(500000000)),
		)

		return after(properMsg)
	}

	// validate mint message was created as intended
	msg := createMsg(func(msg types.MsgMint) types.MsgMint {
		return msg
	})
	codec := initiaappparams.MakeEncodingConfig().Codec
	ac := authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())

	signers, _, err := codec.GetMsgV1Signers(&msg)
	require.NoError(t, err, "can get signers from msg", signers)
	signer, err := ac.BytesToString(signers[0])
	require.NoError(t, err, "can parse signer to string")
	require.Equal(t, len(signers), 1)
	require.Equal(t, signer, addr1.String())

	tests := []struct {
		name       string
		msg        types.MsgMint
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: createMsg(func(msg types.MsgMint) types.MsgMint {
				return msg
			}),
			expectPass: true,
		},
		{
			name: "empty sender",
			msg: createMsg(func(msg types.MsgMint) types.MsgMint {
				msg.Sender = ""
				return msg
			}),
			expectPass: false,
		},
		{
			name: "zero amount",
			msg: createMsg(func(msg types.MsgMint) types.MsgMint {
				msg.Amount = sdk.NewCoin("bitcoin", sdkmath.ZeroInt())
				return msg
			}),
			expectPass: false,
		},
		{
			name: "negative amount",
			msg: createMsg(func(msg types.MsgMint) types.MsgMint {
				msg.Amount.Amount = sdkmath.NewInt(-10000000)
				return msg
			}),
			expectPass: false,
		},
	}

	for _, test := range tests {
		if test.expectPass {
			require.NoError(t, test.msg.Validate(ac), "test: %v", test.name)
		} else {
			require.Error(t, test.msg.Validate(ac), "test: %v", test.name)
		}
	}
}

// TestMsgBurn tests if valid/invalid create denom messages are properly validated/invalidated
func TestMsgBurn(t *testing.T) {
	// generate a private/public key pair and get the respective address
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address())

	// make a proper burn message
	baseMsg := types.NewMsgBurn(
		addr1.String(),
		sdk.NewCoin("bitcoin", sdkmath.NewInt(500000000)),
	)

	// validate burn message was created as intended
	codec := initiaappparams.MakeEncodingConfig().Codec
	ac := authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())

	signers, _, err := codec.GetMsgV1Signers(baseMsg)
	require.NoError(t, err, "can get signers from msg", signers)
	signer, err := ac.BytesToString(signers[0])
	require.NoError(t, err, "can parse signer to string")
	require.Equal(t, len(signers), 1)
	require.Equal(t, signer, addr1.String())

	tests := []struct {
		name       string
		msg        func() *types.MsgBurn
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: func() *types.MsgBurn {
				msg := baseMsg
				return msg
			},
			expectPass: true,
		},
		{
			name: "empty sender",
			msg: func() *types.MsgBurn {
				msg := baseMsg
				msg.Sender = ""
				return msg
			},
			expectPass: false,
		},
		{
			name: "zero amount",
			msg: func() *types.MsgBurn {
				msg := baseMsg
				msg.Amount.Amount = sdkmath.ZeroInt()
				return msg
			},
			expectPass: false,
		},
		{
			name: "negative amount",
			msg: func() *types.MsgBurn {
				msg := baseMsg
				msg.Amount.Amount = sdkmath.NewInt(-10000000)
				return msg
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		if test.expectPass {
			require.NoError(t, test.msg().Validate(ac), "test: %v", test.name)
		} else {
			require.Error(t, test.msg().Validate(ac), "test: %v", test.name)
		}
	}
}

// TestMsgChangeAdmin tests if valid/invalid create denom messages are properly validated/invalidated
func TestMsgChangeAdmin(t *testing.T) {
	// generate a private/public key pair and get the respective address
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address())
	pk2 := ed25519.GenPrivKey().PubKey()
	addr2 := sdk.AccAddress(pk2.Address())
	tokenFactoryDenom := fmt.Sprintf("factory/%s/bitcoin", addr1.String())

	// make a proper changeAdmin message
	baseMsg := types.NewMsgChangeAdmin(
		addr1.String(),
		tokenFactoryDenom,
		addr2.String(),
	)

	// validate changeAdmin message was created as intended
	codec := initiaappparams.MakeEncodingConfig().Codec
	ac := authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())

	signers, _, err := codec.GetMsgV1Signers(baseMsg)
	require.NoError(t, err, "can get signers from msg", signers)
	signer, err := ac.BytesToString(signers[0])
	require.NoError(t, err, "can parse signer to string")
	require.Equal(t, len(signers), 1)
	require.Equal(t, signer, addr1.String())

	tests := []struct {
		name       string
		msg        func() *types.MsgChangeAdmin
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: func() *types.MsgChangeAdmin {
				msg := baseMsg
				return msg
			},
			expectPass: true,
		},
		{
			name: "empty sender",
			msg: func() *types.MsgChangeAdmin {
				msg := baseMsg
				msg.Sender = ""
				return msg
			},
			expectPass: false,
		},
		{
			name: "empty newAdmin",
			msg: func() *types.MsgChangeAdmin {
				msg := baseMsg
				msg.NewAdmin = ""
				return msg
			},
			expectPass: false,
		},
		{
			name: "invalid denom",
			msg: func() *types.MsgChangeAdmin {
				msg := baseMsg
				msg.Denom = "bitcoin"
				return msg
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		if test.expectPass {
			require.NoError(t, test.msg().Validate(ac), "test: %v", test.name)
		} else {
			require.Error(t, test.msg().Validate(ac), "test: %v", test.name)
		}
	}
}

// TestMsgSetDenomMetadata tests if valid/invalid create denom messages are properly validated/invalidated
func TestMsgSetDenomMetadata(t *testing.T) {
	// generate a private/public key pair and get the respective address
	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address())
	tokenFactoryDenom := fmt.Sprintf("factory/%s/bitcoin", addr1.String())
	denomMetadata := banktypes.Metadata{
		Description: "nakamoto",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    tokenFactoryDenom,
				Exponent: 0,
			},
			{
				Denom:    "sats",
				Exponent: 6,
			},
		},
		Display: "sats",
		Base:    tokenFactoryDenom,
		Name:    "bitcoin",
		Symbol:  "BTC",
	}
	invalidDenomMetadata := banktypes.Metadata{
		Description: "nakamoto",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    "bitcoin",
				Exponent: 0,
			},
			{
				Denom:    "sats",
				Exponent: 6,
			},
		},
		Display: "sats",
		Base:    "bitcoin",
		Name:    "bitcoin",
		Symbol:  "BTC",
	}

	// make a proper setDenomMetadata message
	baseMsg := types.NewMsgSetDenomMetadata(
		addr1.String(),
		denomMetadata,
	)

	// validate setDenomMetadata message was created as intended
	codec := initiaappparams.MakeEncodingConfig().Codec
	ac := authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())

	signers, _, err := codec.GetMsgV1Signers(baseMsg)
	require.NoError(t, err, "can get signers from msg", signers)
	signer, err := ac.BytesToString(signers[0])
	require.NoError(t, err, "can parse signer to string")
	require.Equal(t, len(signers), 1)
	require.Equal(t, signer, addr1.String())

	tests := []struct {
		name       string
		msg        func() *types.MsgSetDenomMetadata
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: func() *types.MsgSetDenomMetadata {
				msg := baseMsg
				return msg
			},
			expectPass: true,
		},
		{
			name: "empty sender",
			msg: func() *types.MsgSetDenomMetadata {
				msg := baseMsg
				msg.Sender = ""
				return msg
			},
			expectPass: false,
		},
		{
			name: "invalid metadata",
			msg: func() *types.MsgSetDenomMetadata {
				msg := baseMsg
				msg.Metadata.Name = ""
				return msg
			},

			expectPass: false,
		},
		{
			name: "invalid base",
			msg: func() *types.MsgSetDenomMetadata {
				msg := baseMsg
				msg.Metadata = invalidDenomMetadata
				return msg
			},
			expectPass: false,
		},
	}

	for _, test := range tests {
		if test.expectPass {
			require.NoError(t, test.msg().Validate(ac), "test: %v", test.name)
		} else {
			require.Error(t, test.msg().Validate(ac), "test: %v", test.name)
		}
	}
}
