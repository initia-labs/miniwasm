package keeper

import (
	"context"

	"net/url"

	"github.com/initia-labs/miniwasm/x/tokenfactory/types"
)

type Querier struct {
	*Keeper
}

var _ types.QueryServer = Querier{}

func (q Querier) Params(ctx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	params := q.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

func (q Querier) DenomAuthorityMetadata(ctx context.Context, req *types.QueryDenomAuthorityMetadataRequest) (*types.QueryDenomAuthorityMetadataResponse, error) {
	decodedDenom, err := url.QueryUnescape(req.Denom)
	if err == nil {
		req.Denom = decodedDenom
	}
	authorityMetadata, err := q.GetAuthorityMetadata(ctx, req.GetDenom())
	if err != nil {
		return nil, err
	}

	return &types.QueryDenomAuthorityMetadataResponse{AuthorityMetadata: authorityMetadata}, nil
}

func (q Querier) DenomsFromCreator(ctx context.Context, req *types.QueryDenomsFromCreatorRequest) (*types.QueryDenomsFromCreatorResponse, error) {
	denoms := q.getDenomsFromCreator(ctx, req.GetCreator())
	return &types.QueryDenomsFromCreatorResponse{Denoms: denoms}, nil
}

func (q Querier) BeforeSendHookAddress(ctx context.Context, req *types.QueryBeforeSendHookAddressRequest) (*types.QueryBeforeSendHookAddressResponse, error) {
	decodedDenom, err := url.QueryUnescape(req.Denom)
	if err == nil {
		req.Denom = decodedDenom
	}

	cosmwasmAddress := q.GetBeforeSendHook(ctx, req.GetDenom())

	return &types.QueryBeforeSendHookAddressResponse{CosmwasmAddress: cosmwasmAddress}, nil
}
