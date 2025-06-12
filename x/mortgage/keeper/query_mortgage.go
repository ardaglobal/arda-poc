package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/ardaglobal/arda-poc/x/mortgage/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) MortgageAll(ctx context.Context, req *types.QueryAllMortgageRequest) (*types.QueryAllMortgageResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var mortgages []types.Mortgage

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	mortgageStore := prefix.NewStore(store, types.KeyPrefix(types.MortgageKeyPrefix))

	pageRes, err := query.Paginate(mortgageStore, req.Pagination, func(key []byte, value []byte) error {
		var mortgage types.Mortgage
		if err := k.cdc.Unmarshal(value, &mortgage); err != nil {
			return err
		}

		mortgages = append(mortgages, mortgage)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllMortgageResponse{Mortgage: mortgages, Pagination: pageRes}, nil
}

func (k Keeper) Mortgage(ctx context.Context, req *types.QueryGetMortgageRequest) (*types.QueryGetMortgageResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	val, found := k.GetMortgage(
		ctx,
		req.Index,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetMortgageResponse{Mortgage: val}, nil
}
