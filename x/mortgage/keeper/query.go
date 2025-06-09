package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/ardaglobal/arda-poc/x/mortgage/types"
)

func (k Keeper) MortgageAll(goCtx context.Context, req *types.QueryAllMortgageRequest) (*types.QueryAllMortgageResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	// Example: no pagination for now
	mortgages := []types.Mortgage{}
	store := k.storeService.OpenKVStore(ctx)
	prefix := []byte(types.KeyPrefixMortgage)
	iter, err := store.Iterator(prefix, nil)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var m types.Mortgage
		k.cdc.MustUnmarshal(iter.Value(), &m)
		mortgages = append(mortgages, m)
	}
	res := make([]*types.Mortgage, len(mortgages))
	for i := range mortgages {
		res[i] = &mortgages[i]
	}
	return &types.QueryAllMortgageResponse{Mortgages: res, Pagination: &query.PageResponse{Total: uint64(len(res))}}, nil
}

func (k Keeper) Mortgage(goCtx context.Context, req *types.QueryGetMortgageRequest) (*types.QueryGetMortgageResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	m, found := k.GetMortgage(ctx, req.Index)
	if !found {
		return nil, status.Error(codes.NotFound, "mortgage not found")
	}
	return &types.QueryGetMortgageResponse{Mortgage: &m}, nil
}

func (k Keeper) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	params := k.GetParams(ctx)
	return &types.QueryParamsResponse{Params: params}, nil
}
