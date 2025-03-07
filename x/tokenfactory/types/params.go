package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Parameter store keys.
var (
	KeyDenomCreationFee        = []byte("DenomCreationFee")
	KeyDenomCreationGasConsume = []byte("DenomCreationGasConsume")

	// chosen as an arbitrary large number, less than the max_gas_wanted_per_tx in config.
	DefaultCreationGasFee = 1_000_000
)

func NewParams(denomCreationFee sdk.Coins, denomCreationGasConsume uint64) Params {
	return Params{
		DenomCreationFee:        denomCreationFee,
		DenomCreationGasConsume: denomCreationGasConsume,
	}
}

// default gamm module parameters.
func DefaultParams() Params {
	return Params{
		// For choice, see: https://github.com/osmosis-labs/osmosis/pull/4983
		DenomCreationFee:        sdk.NewCoins(), // used to be 10 OSMO at launch.
		DenomCreationGasConsume: uint64(DefaultCreationGasFee),
	}
}

// validate params.
func (p Params) Validate() error {
	if err := validateDenomCreationFee(p.DenomCreationFee); err != nil {
		return err
	}

	if err := validateDenomCreationGasConsume(p.DenomCreationGasConsume); err != nil {
		return err
	}

	return nil
}

func validateDenomCreationFee(i any) error {
	v, ok := i.(sdk.Coins)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.Validate() != nil {
		return fmt.Errorf("invalid denom creation fee: %+v", i)
	}

	return nil
}

func validateDenomCreationGasConsume(i any) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}
