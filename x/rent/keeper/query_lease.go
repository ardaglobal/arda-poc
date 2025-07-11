package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	"github.com/ardaglobal/arda-poc/x/rent/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) LeaseAll(ctx context.Context, req *types.QueryAllLeaseRequest) (*types.QueryAllLeaseResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var leases []types.Lease

	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	leaseStore := prefix.NewStore(store, types.KeyPrefix(types.LeaseKey))

	pageRes, err := query.Paginate(leaseStore, req.Pagination, func(key []byte, value []byte) error {
		var lease types.Lease
		if err := k.cdc.Unmarshal(value, &lease); err != nil {
			return err
		}

		leases = append(leases, lease)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllLeaseResponse{Lease: leases, Pagination: pageRes}, nil
}

func (k Keeper) Lease(ctx context.Context, req *types.QueryGetLeaseRequest) (*types.QueryGetLeaseResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	lease, found := k.GetLease(ctx, req.Id)
	if !found {
		return nil, sdkerrors.ErrKeyNotFound
	}

	return &types.QueryGetLeaseResponse{Lease: lease}, nil
}
