package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	propertyTypes "github.com/ardaglobal/arda-poc/x/property/types"
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
	var paginatedProperties []*propertyTypes.Property
	var pageRes *query.PageResponse
	
	if req.Pagination != nil {
		// Use proper pagination with default limit
		limit := req.Pagination.Limit
		if limit == 0 {
			limit = 100 // Default limit
		}
		if limit > 1000 {
			limit = 1000 // Max limit
		}
		
		offset := req.Pagination.Offset
		if offset > uint64(len(propertyPtrs)) {
			offset = uint64(len(propertyPtrs))
		}
		
		start := int(offset)
		end := start + int(limit)
		if end > len(propertyPtrs) {
			end = len(propertyPtrs)
		}
		
		if start < end {
			paginatedProperties = propertyPtrs[start:end]
		} else {
			paginatedProperties = []*propertyTypes.Property{}
		}
		
		// Create pagination response with next key for cursor-based pagination
		pageRes = &query.PageResponse{
			Total: uint64(len(propertyPtrs)),
		}
		
		if req.Pagination.CountTotal {
			pageRes.Total = uint64(len(propertyPtrs))
		}
		
		// Set next key if there are more results
		if end < len(propertyPtrs) {
			pageRes.NextKey = []byte(fmt.Sprintf("%d", end))
		}
	} else {
		// No pagination requested, return all (with reasonable limit)
		if len(propertyPtrs) > 100 {
			paginatedProperties = propertyPtrs[:100]
			pageRes = &query.PageResponse{
				Total:   uint64(len(propertyPtrs)),
				NextKey: []byte("100"),
			}
		} else {
			paginatedProperties = propertyPtrs
			pageRes = &query.PageResponse{
				Total: uint64(len(propertyPtrs)),
			}
		}
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
