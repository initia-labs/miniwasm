package types

import (
	"cosmossdk.io/core/address"
	errorsmod "cosmossdk.io/errors"
)

// this line is used by starport scaffolding # genesis/types/import

// DefaultIndex is the default capability global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default Capability genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:        DefaultParams(),
		FactoryDenoms: []GenesisDenom{},
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate(ac address.Codec) error {
	err := gs.Params.Validate()
	if err != nil {
		return err
	}

	seenDenoms := map[string]bool{}

	for _, denom := range gs.GetFactoryDenoms() {
		if seenDenoms[denom.GetDenom()] {
			return errorsmod.Wrapf(ErrInvalidGenesis, "duplicate denom: %s", denom.GetDenom())
		}
		seenDenoms[denom.GetDenom()] = true

		_, _, err := DeconstructDenom(ac, denom.GetDenom())
		if err != nil {
			return err
		}

		if denom.AuthorityMetadata.Admin != "" {
			_, err = ac.StringToBytes(denom.AuthorityMetadata.Admin)
			if err != nil {
				return errorsmod.Wrapf(ErrInvalidAuthorityMetadata, "Invalid admin address (%s)", err)
			}
		}
	}

	return nil
}
