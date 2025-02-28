package types_test

import (
	fmt "fmt"
	"testing"

	"github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authcodec "github.com/cosmos/cosmos-sdk/x/auth/codec"

	"github.com/initia-labs/miniwasm/x/tokenfactory/types"
)

func TestDeconstructDenom(t *testing.T) {
	ac := authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
	bytesAddr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	creator, err := ac.BytesToString(bytesAddr)
	require.NoError(t, err, "address bytes encoding error")

	for _, tc := range []struct {
		desc             string
		denom            string
		expectedSubdenom string
		err              error
	}{
		{
			desc:  "empty is invalid",
			denom: "",
			err:   types.ErrInvalidDenom,
		},
		{
			desc:             "normal",
			denom:            fmt.Sprintf("factory/%s/bitcoin", creator),
			expectedSubdenom: "bitcoin",
		},
		{
			desc:             "multiple slashes in subdenom",
			denom:            fmt.Sprintf("factory/%s/bitcoin/1", creator),
			expectedSubdenom: "bitcoin/1",
		},
		{
			desc:             "no subdenom",
			denom:            fmt.Sprintf("factory/%s/", creator),
			expectedSubdenom: "",
		},
		{
			desc:  "incorrect prefix",
			denom: fmt.Sprintf("ibc/%s/bitcoin", creator),
			err:   types.ErrInvalidDenom,
		},
		{
			desc:             "subdenom of only slashes",
			denom:            fmt.Sprintf("factory/%s/////", creator),
			expectedSubdenom: "////",
		},
		{
			desc:  "too long name",
			denom: fmt.Sprintf("factory/%s/adsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsf", creator),
			err:   types.ErrInvalidDenom,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			expectedCreator := creator
			creator, subdenom, err := types.DeconstructDenom(ac, tc.denom)
			if tc.err != nil {
				require.ErrorContains(t, err, tc.err.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, expectedCreator, creator)
				require.Equal(t, tc.expectedSubdenom, subdenom)
			}
		})
	}
}

func TestGetTokenDenom(t *testing.T) {
	ac := authcodec.NewBech32Codec(sdk.GetConfig().GetBech32AccountAddrPrefix())
	bytesAddr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	creator, err := ac.BytesToString(bytesAddr)
	require.NoError(t, err, "address bytes encoding error")

	for _, tc := range []struct {
		desc     string
		creator  string
		subdenom string
		valid    bool
	}{
		{
			desc:     "normal",
			creator:  creator,
			subdenom: "bitcoin",
			valid:    true,
		},
		{
			desc:     "multiple slashes in subdenom",
			creator:  creator,
			subdenom: "bitcoin/1",
			valid:    true,
		},
		{
			desc:     "no subdenom",
			creator:  creator,
			subdenom: "",
			valid:    true,
		},
		{
			desc:     "subdenom of only slashes",
			creator:  creator,
			subdenom: "/////",
			valid:    true,
		},
		{
			desc:     "too long name",
			creator:  creator,
			subdenom: "adsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsfadsf",
			valid:    false,
		},
		{
			desc:     "subdenom is exactly max length",
			creator:  creator,
			subdenom: "bitcoinfsadfsdfeadfsafwefsefsefsdfsdafasefsf",
			valid:    true,
		},
		{
			desc:     "creator is exactly max length",
			creator:  "init1t7egva48prqmzl59x5ngv4zx0dtrwewc9m7z44jhgjhgkhjklhkjhkjhgjhgjgjghelugt",
			subdenom: "bitcoin",
			valid:    true,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			_, err := types.GetTokenDenom(tc.creator, tc.subdenom)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
