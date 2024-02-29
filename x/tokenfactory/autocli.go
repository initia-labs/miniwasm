package tokenfactory

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	tokenfactoryv1beta1 "github.com/initia-labs/miniwasm/api/miniwasm/tokenfactory/v1beta1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: tokenfactoryv1beta1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "DenomAuthorityMetadata",
					Use:       "denom-authority-metadata [denom]",
					Short:     "Get the authority metadata for a specific denom",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "denom"},
					},
				},
				{
					RpcMethod: "DenomsFromCreator",
					Use:       "denoms-from-creator [creator]",
					Short:     "Returns a list of all tokens created by a specific creator address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "creator"},
					},
				},
				{
					RpcMethod: "BeforeSendHookAddress",
					Use:       "before_send_hook [denom]",
					Short:     "Get the BeforeSend hook for a specific denom",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "denom"},
					},
				},
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Returns the tokenfactory module's parameters",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: tokenfactoryv1beta1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "CreateDenom",
					Use:       "create-denom [sender] [sub-denom]",
					Short:     "create a new denom from an account.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender"},
						{ProtoField: "subdenom"},
					},
				},
				{
					RpcMethod: "Mint",
					Use:       "mint [sender] [amount] [mint-to-address]",
					Short:     "Mint a denom to an address. Must have admin authority to do so.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender"},
						{ProtoField: "amount"},
						{ProtoField: "mintToAddress"},
					},
				},
				{
					RpcMethod: "Burn",
					Use:       "burn [sender] [amount] [burn-from-address]",
					Short:     "Burn tokens from an address. Must have admin authority to do so.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender"},
						{ProtoField: "amount"},
						{ProtoField: "burnFromAddress"},
					},
				},
				{
					RpcMethod: "ChangeAdmin",
					Use:       "change-admin [sender] [denom] [new-admin]",
					Short:     "Changes the admin address for a factory-created denom. Must have admin authority to do so.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender"},
						{ProtoField: "denom"},
						{ProtoField: "new_admin"},
					},
				},
				{
					RpcMethod: "SetDenomMetadata",
					Use:       "set-denom-metadata [sender] [denom] [cosmwasm-address]",
					Short:     "Set a denom metadata",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender"},
						{ProtoField: "denom"},
						{ProtoField: "cosmwasm_address"},
					},
				},
				{
					RpcMethod: "SetBeforeSendHook",
					Use:       "set-beforesend-hook [denom] [cosmwasm-address]",
					Short:     "Set a cosmwasm contract to be the beforesend hook for a factory-created denom. Must have admin authority to do so.",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender"},
						{ProtoField: "subdenom"},
					},
				},
				{
					RpcMethod: "ForceTransfer",
					Use:       "force-transfer [sender] [amount] [transfer-from-address] [transfer-to-address]",
					Short:     "Transfer a factory-crated denom",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "sender"},
						{ProtoField: "amount"},
						{ProtoField: "transferFromAddress"},
						{ProtoField: "transferToAddress"},
					},
				},
			},
			EnhanceCustomCommand: false, // use custom commands only until v0.51
		},
	}
}
