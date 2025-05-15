package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	propertyTypes "arda/x/property/types"
)

func (k Keeper) PropertyAll(goCtx context.Context, req *propertyTypes.QueryAllPropertyRequest) (*propertyTypes.QueryAllPropertyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	// Use GetAllSubmissions to get all submissions
	allProperties, err := k.GetAllProperties(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Convert to pointers for response type
	propertyPtrs := make([]*propertyTypes.Property, len(allProperties))
	for i, prop := range allProperties {
		propertyCopy := prop // Create a copy to avoid reference issues
		propertyPtrs[i] = &propertyCopy
	}

	// Apply pagination if provided
	start, end := 0, len(propertyPtrs)
	if req.Pagination != nil {
		start = int(req.Pagination.Offset)
		if start > len(propertyPtrs) {
			start = len(propertyPtrs)
		}

		if req.Pagination.Limit > 0 && start+int(req.Pagination.Limit) < len(propertyPtrs) {
			end = start + int(req.Pagination.Limit)
		}
	}

	// Slice the submissions based on pagination
	paginatedProperties := propertyPtrs
	if start < end {
		paginatedProperties = propertyPtrs[start:end]
	} else {
		paginatedProperties = []*propertyTypes.Property{}
	}

	// Create pagination response
	pageRes := &query.PageResponse{
		Total: uint64(len(propertyPtrs)),
	}
	if req.Pagination != nil && req.Pagination.CountTotal {
		pageRes.Total = uint64(len(propertyPtrs))
	}

	return &propertyTypes.QueryAllPropertyResponse{
		Properties: paginatedProperties,
		Pagination: pageRes,
	}, nil
}

func (k Keeper) Property(goCtx context.Context, req *propertyTypes.QueryGetPropertyRequest) (*propertyTypes.QueryGetPropertyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	property, found := k.GetProperty(ctx, req.Index)
	if !found {
		return nil, status.Error(codes.NotFound, "property not found")
	}

	return &propertyTypes.QueryGetPropertyResponse{Property: &property}, nil
}
