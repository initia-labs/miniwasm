package types

import (
	"cosmossdk.io/core/address"

	sdkmath "cosmossdk.io/math"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// constants
const (
	TypeMsgCreateDenom       = "create_denom"
	TypeMsgMint              = "tf_mint"
	TypeMsgBurn              = "tf_burn"
	TypeMsgForceTransfer     = "force_transfer"
	TypeMsgChangeAdmin       = "change_admin"
	TypeMsgSetDenomMetadata  = "set_denom_metadata"
	TypeMsgSetBeforeSendHook = "set_before_send_hook"
)

var (
	_ sdk.Msg = &MsgCreateDenom{}
	_ sdk.Msg = &MsgMint{}
	_ sdk.Msg = &MsgBurn{}
	_ sdk.Msg = &MsgChangeAdmin{}
	_ sdk.Msg = &MsgSetDenomMetadata{}
	_ sdk.Msg = &MsgSetBeforeSendHook{}
)

// NewMsgCreateDenom creates a msg to create a new denom
func NewMsgCreateDenom(sender, subdenom string) *MsgCreateDenom {
	return &MsgCreateDenom{
		Sender:   sender,
		Subdenom: subdenom,
	}
}

func (m MsgCreateDenom) Validate(accAddrCodec address.Codec) error {
	if addr, err := accAddrCodec.StringToBytes(m.Sender); err != nil {
		return err
	} else if len(addr) == 0 {
		return ErrEmptySender
	}

	if _, err := GetTokenDenom(m.Sender, m.Subdenom); err != nil {
		return ErrInvalidDenom
	}

	return nil
}

// NewMsgMint creates a message to mint tokens
func NewMsgMint(sender string, amount sdk.Coin) *MsgMint {
	return &MsgMint{
		Sender: sender,
		Amount: amount,
	}
}

func NewMsgMintTo(sender string, amount sdk.Coin, mintToAddress string) *MsgMint {
	return &MsgMint{
		Sender:        sender,
		Amount:        amount,
		MintToAddress: mintToAddress,
	}
}

func (m MsgMint) Validate(accAddrCodec address.Codec) error {
	if addr, err := accAddrCodec.StringToBytes(m.Sender); err != nil {
		return err
	} else if len(addr) == 0 {
		return ErrEmptySender
	}

	if !m.Amount.IsValid() || m.Amount.Amount.Equal(sdkmath.ZeroInt()) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, m.Amount.String())
	}

	return nil
}

// NewMsgBurn creates a message to burn tokens
func NewMsgBurn(sender string, amount sdk.Coin) *MsgBurn {
	return &MsgBurn{
		Sender: sender,
		Amount: amount,
	}
}

func (m MsgBurn) Validate(accAddrCodec address.Codec) error {
	if addr, err := accAddrCodec.StringToBytes(m.Sender); err != nil {
		return err
	} else if len(addr) == 0 {
		return ErrEmptySender
	}

	if !m.Amount.IsValid() || m.Amount.Amount.Equal(sdkmath.ZeroInt()) {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, m.Amount.String())
	}

	return nil
}

// NewMsgChangeAdmin creates a message to burn tokens
func NewMsgChangeAdmin(sender, denom, newAdmin string) *MsgChangeAdmin {
	return &MsgChangeAdmin{
		Sender:   sender,
		Denom:    denom,
		NewAdmin: newAdmin,
	}
}

func (m MsgChangeAdmin) Validate(accAddrCodec address.Codec) error {
	if _, err := accAddrCodec.StringToBytes(m.Sender); err != nil {
		return err
	}

	// allow empty address
	if len(m.NewAdmin) > 0 {
		if _, err := accAddrCodec.StringToBytes(m.NewAdmin); err != nil {
			return err
		}
	}

	if _, _, err := DeconstructDenom(accAddrCodec, m.Denom); err != nil {
		return err
	}

	return nil
}

// NewMsgChangeAdmin creates a message to burn tokens
func NewMsgSetDenomMetadata(sender string, metadata banktypes.Metadata) *MsgSetDenomMetadata {
	return &MsgSetDenomMetadata{
		Sender:   sender,
		Metadata: metadata,
	}
}

func (m MsgSetDenomMetadata) Validate(accAddrCodec address.Codec) error {
	if addr, err := accAddrCodec.StringToBytes(m.Sender); err != nil {
		return err
	} else if len(addr) == 0 {
		return ErrEmptySender
	}

	if err := m.Metadata.Validate(); err != nil {
		return err
	}

	if _, _, err := DeconstructDenom(accAddrCodec, m.Metadata.Base); err != nil {
		return err
	}
	return nil
}

// NewMsgSetBeforeSendHook creates a message to set a new before send hook
func NewMsgSetBeforeSendHook(sender string, denom string, cosmwasmAddress string) *MsgSetBeforeSendHook {
	return &MsgSetBeforeSendHook{
		Sender:          sender,
		Denom:           denom,
		CosmwasmAddress: cosmwasmAddress,
	}
}

func (m MsgSetBeforeSendHook) Validate(accAddrCodec address.Codec) error {
	if addr, err := accAddrCodec.StringToBytes(m.Sender); err != nil {
		return err
	} else if len(addr) == 0 {
		return ErrEmptySender
	}

	if _, _, err := DeconstructDenom(accAddrCodec, m.Denom); err != nil {
		return ErrInvalidDenom
	}
	return nil
}
