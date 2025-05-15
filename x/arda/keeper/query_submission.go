package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	ardaTypes "github.com/ardaglobal/arda-poc/x/arda/types"
)

func (k Keeper) SubmissionAll(goCtx context.Context, req *ardaTypes.QueryAllSubmissionRequest) (*ardaTypes.QueryAllSubmissionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Use GetAllSubmissions to get all submissions
	allSubmissions, err := k.GetAllSubmissions(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convert to pointers for response type
	submissionPtrs := make([]*ardaTypes.Submission, len(allSubmissions))
	for i, sub := range allSubmissions {
		submissionCopy := sub // Create a copy to avoid reference issues
		submissionPtrs[i] = &submissionCopy
	}

	// Apply pagination if provided
	start, end := 0, len(submissionPtrs)
	if req.Pagination != nil {
		start = int(req.Pagination.Offset)
		if start > len(submissionPtrs) {
			start = len(submissionPtrs)
		}

		if req.Pagination.Limit > 0 && start+int(req.Pagination.Limit) < len(submissionPtrs) {
			end = start + int(req.Pagination.Limit)
		}
	}

	// Slice the submissions based on pagination
	paginatedSubmissions := submissionPtrs
	if start < end {
		paginatedSubmissions = submissionPtrs[start:end]
	} else {
		paginatedSubmissions = []*ardaTypes.Submission{}
	}

	// Create pagination response
	pageRes := &query.PageResponse{
		Total: uint64(len(submissionPtrs)),
	}
	if req.Pagination != nil && req.Pagination.CountTotal {
		pageRes.Total = uint64(len(submissionPtrs))
	}

	return &ardaTypes.QueryAllSubmissionResponse{
		Submission: paginatedSubmissions,
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
