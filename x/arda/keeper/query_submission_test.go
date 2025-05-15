package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"

	keepertest "github.com/ardaglobal/arda-poc/testutil/keeper"
	"github.com/ardaglobal/arda-poc/x/arda/types"
)

func TestSubmissionAllQuery(t *testing.T) {
	k, ctx := keepertest.ArdaKeeper(t)

	// Create query server using the keeper instance
	queryServer := k.SubmissionAll

	// Add submissions to the store
	submissions := []types.Submission{
		{Creator: "addr1", Region: "dubai", Hash: "hash1", Valid: "true"},
		{Creator: "addr2", Region: "dubai", Hash: "hash2", Valid: "false"},
		{Creator: "addr3", Region: "singapore", Hash: "hash3", Valid: "true"},
		{Creator: "addr4", Region: "london", Hash: "hash4", Valid: "true"},
		{Creator: "addr5", Region: "singapore", Hash: "hash5", Valid: "false"},
	}

	for _, sub := range submissions {
		k.AppendSubmission(ctx, sub)
	}

	// Test without pagination (get all)
	resp, err := queryServer(ctx, &types.QueryAllSubmissionRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(len(submissions)), resp.Pagination.Total)
	require.Len(t, resp.Submission, len(submissions))

	// Test with pagination (limit = 2)
	resp, err = queryServer(ctx, &types.QueryAllSubmissionRequest{
		Pagination: &query.PageRequest{
			Limit:      2,
			CountTotal: true,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(len(submissions)), resp.Pagination.Total)
	require.Len(t, resp.Submission, 2)

	// Test with pagination (offset = 2, limit = 2)
	resp, err = queryServer(ctx, &types.QueryAllSubmissionRequest{
		Pagination: &query.PageRequest{
			Offset:     2,
			Limit:      2,
			CountTotal: true,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(len(submissions)), resp.Pagination.Total)
	require.Len(t, resp.Submission, 2)

	// Test with offset beyond available items
	resp, err = queryServer(ctx, &types.QueryAllSubmissionRequest{
		Pagination: &query.PageRequest{
			Offset:     10,
			CountTotal: true,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, uint64(len(submissions)), resp.Pagination.Total)
	require.Empty(t, resp.Submission)
}

func TestSubmissionQuery(t *testing.T) {
	k, ctx := keepertest.ArdaKeeper(t)

	// Create query server for single submission queries
	queryServer := k.Submission

	// Add submissions to the store
	submissions := []types.Submission{
		{Creator: "addr1", Region: "dubai", Hash: "hash1", Valid: "true"},
		{Creator: "addr2", Region: "singapore", Hash: "hash2", Valid: "false"},
		{Creator: "addr3", Region: "london", Hash: "hash3", Valid: "true"},
	}

	var ids []uint64
	for _, sub := range submissions {
		id := k.AppendSubmission(ctx, sub)
		ids = append(ids, id)
	}

	// Test getting each submission by ID
	for i, id := range ids {
		resp, err := queryServer(ctx, &types.QueryGetSubmissionRequest{Id: id})
		require.NoError(t, err)
		require.NotNil(t, resp)
		require.NotNil(t, resp.Submission)

		sub := resp.Submission
		require.Equal(t, submissions[i].Creator, sub.Creator)
		require.Equal(t, submissions[i].Region, sub.Region)
		require.Equal(t, submissions[i].Hash, sub.Hash)
		require.Equal(t, submissions[i].Valid, sub.Valid)
	}

	// Test getting a non-existent submission
	nonExistentID := uint64(999)
	_, err := queryServer(ctx, &types.QueryGetSubmissionRequest{Id: nonExistentID})
	require.Error(t, err)
	require.Contains(t, err.Error(), "submission not found")
}
