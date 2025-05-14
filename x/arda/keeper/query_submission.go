package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	ardaTypes "arda/x/arda/types"

	storetypes "cosmossdk.io/store/types"
)

func (k Keeper) SubmissionAll(goCtx context.Context, req *ardaTypes.QueryAllSubmissionRequest) (*ardaTypes.QueryAllSubmissionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var submissions []*ardaTypes.Submission
	ctx := sdk.UnwrapSDKContext(goCtx)

	kvStore := k.storeService.OpenKVStore(ctx).(storetypes.KVStore)

	pageRes, err := query.Paginate(kvStore, req.Pagination, func(key []byte, value []byte) error {
		var submission ardaTypes.Submission
		if err := k.cdc.Unmarshal(value, &submission); err != nil {
			return err
		}
		submissions = append(submissions, &submission)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &ardaTypes.QueryAllSubmissionResponse{
		Submission: submissions,
		Pagination: pageRes,
	}, nil
}
func (k Keeper) Submission(goCtx context.Context, req *ardaTypes.QueryGetSubmissionRequest) (*ardaTypes.QueryGetSubmissionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	submission, found := k.GetSubmission(ctx, req.Id)
	if !found {
		return nil, status.Error(codes.NotFound, "submission not found")
	}

	return &ardaTypes.QueryGetSubmissionResponse{Submission: &submission}, nil
}
