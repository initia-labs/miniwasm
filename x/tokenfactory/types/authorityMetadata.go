package types

import (
	"cosmossdk.io/core/address"
)

func (metadata DenomAuthorityMetadata) Validate(ac address.Codec) error {
	if metadata.Admin != "" {
		_, err := ac.StringToBytes(metadata.Admin)
		if err != nil {
			return err
		}
	}
	return nil
}
