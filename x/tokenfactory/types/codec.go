package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	// this line is used by starport scaffolding # 1
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgCreateDenom{}, "tokenfactory/MsgCreateDenom")
	legacy.RegisterAminoMsg(cdc, &MsgMint{}, "tokenfactory/MsgMint")
	legacy.RegisterAminoMsg(cdc, &MsgBurn{}, "tokenfactory/MsgBurn")
	legacy.RegisterAminoMsg(cdc, &MsgChangeAdmin{}, "tokenfactory/MsgChangeAdmin")
	legacy.RegisterAminoMsg(cdc, &MsgSetDenomMetadata{}, "tokenfactory/MsgSetDenomMetadata")
	legacy.RegisterAminoMsg(cdc, &MsgSetBeforeSendHook{}, "tokenfactory/MsgSetBeforeSendHook")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "tokenfactory/MsgUpdateParams")
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgCreateDenom{},
		&MsgMint{},
		&MsgBurn{},
		&MsgChangeAdmin{},
		&MsgSetDenomMetadata{},
		&MsgSetBeforeSendHook{},
		&MsgUpdateParams{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
