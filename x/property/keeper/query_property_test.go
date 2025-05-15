package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"

	keepertest "arda/testutil/keeper"
	"arda/x/property/types"
)

func TestPropertyAllQuery(t *testing.T) {
	k, ctx := keepertest.PropertyKeeper(t)

	// Create query server using the keeper instance
	queryServer := k.PropertyAll

	// Add properties to the store
	properties := []types.Property{
		{Index: "addr1", Region: "dubai", Value: 100},
		{Index: "addr2", Region: "dubai", Value: 200},
		{Index: "addr3", Region: "singapore", Value: 300},
		{Index: "addr4", Region: "london", Value: 400},
		{Index: "addr5", Region: "singapore", Value: 500},
	}

	for _, property := range properties {
		k.SetProperty(ctx, property)
	}

	// Test without pagination (get all)
	resp, err := queryServer(ctx, &types.QueryAllPropertyRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(len(properties)), resp.Pagination.Total)
	require.Len(t, resp.Property, len(properties))

	// Test with pagination (limit = 2)
	resp, err = queryServer(ctx, &types.QueryAllPropertyRequest{
		Pagination: &query.PageRequest{
			Limit:      2,
			CountTotal: true,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(len(properties)), resp.Pagination.Total)
	require.Len(t, resp.Property, 2)

	// Test with pagination (offset = 2, limit = 2)
	resp, err = queryServer(ctx, &types.QueryAllPropertyRequest{
		Pagination: &query.PageRequest{
			Offset:     2,
			Limit:      2,
			CountTotal: true,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(len(properties)), resp.Pagination.Total)
	require.Len(t, resp.Property, 2)

	// Test with offset beyond available items
	resp, err = queryServer(ctx, &types.QueryAllPropertyRequest{
		Pagination: &query.PageRequest{
			Offset:     10,
			CountTotal: true,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(len(properties)), resp.Pagination.Total)
	require.Empty(t, resp.Property)
}

func TestPropertyQuery(t *testing.T) {
	k, ctx := keepertest.ArdaKeeper(t)

	// Create query server for single submission queries
	queryServer := k.Property

	// Add submissions to the store
	properties := []types.Property{
		{Index: "addr1", Region: "dubai", Value: 100},
		{Index: "addr2", Region: "singapore", Value: 200},
		{Index: "addr3", Region: "london", Value: 300},
	}

	var ids []uint64
	for _, property := range properties {
		id := k.SetProperty(ctx, property)
		ids = append(ids, id)
	}

	// Test getting each submission by ID
	for i, id := range ids {
		resp, err := queryServer(ctx, &types.QueryGetPropertyRequest{Index: id})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Property)

		property := resp.Property
		require.Equal(t, properties[i].Index, property.Index)
		require.Equal(t, properties[i].Region, property.Region)
		require.Equal(t, properties[i].Value, property.Value)
		require.Equal(t, properties[i].Owners, property.Owners)
	}

	// Test getting a non-existent submission
	nonExistentID := uint64(999)
	_, err := queryServer(ctx, &types.QueryGetPropertyRequest{Index: nonExistentID})
	require.Error(t, err)
	require.Contains(t, err.Error(), "property not found")
}
