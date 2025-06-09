package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/ardaglobal/arda-poc/testutil/keeper"
	"github.com/ardaglobal/arda-poc/x/arda/types"
)

func TestAppendAndGetSubmission(t *testing.T) {
	k, ctx := keepertest.ArdaKeeper(t)

	submissions := []types.Submission{
		{Creator: "addr1", Region: "dubai", Hash: "hash1", Valid: "true"},
		{Creator: "addr2", Region: "dubai", Hash: "hash2", Valid: "false"},
	}

	for i, sub := range submissions {
		id := k.AppendSubmission(ctx, sub)
		require.Equal(t, uint64(i), id)

		got, found := k.GetSubmission(ctx, id)
		require.True(t, found)
		require.Equal(t, sub.Creator, got.Creator)
		require.Equal(t, sub.Region, got.Region)
		require.Equal(t, sub.Hash, got.Hash)
		require.Equal(t, sub.Valid, got.Valid)
	}

	// Test not found
	_, found := k.GetSubmission(ctx, 999)
	require.False(t, found)

	// Test GetAllSubmissions
	allSubmissions, err := k.GetAllSubmissions(ctx)
	require.NoError(t, err)
	require.Len(t, allSubmissions, len(submissions))

	// Verify each submission
	for i, sub := range submissions {
		require.Equal(t, sub.Creator, allSubmissions[i].Creator)
		require.Equal(t, sub.Region, allSubmissions[i].Region)
		require.Equal(t, sub.Hash, allSubmissions[i].Hash)
		require.Equal(t, sub.Valid, allSubmissions[i].Valid)
	}
}

func TestGetAllSubmissions(t *testing.T) {
	k, ctx := keepertest.ArdaKeeper(t)

	// Empty at start
	emptySubmissions, err := k.GetAllSubmissions(ctx)
	require.NoError(t, err)
	require.Empty(t, emptySubmissions)

	// Add multiple submissions from different creators and regions
	submissions := []types.Submission{
		{Creator: "addr1", Region: "dubai", Hash: "hash1", Valid: "true"},
		{Creator: "addr2", Region: "dubai", Hash: "hash2", Valid: "false"},
		{Creator: "addr3", Region: "singapore", Hash: "hash3", Valid: "true"},
		{Creator: "addr4", Region: "london", Hash: "hash4", Valid: "true"},
		{Creator: "addr5", Region: "singapore", Hash: "hash5", Valid: "false"},
	}

	// Add all submissions to the store
	var ids []uint64
	for _, sub := range submissions {
		id := k.AppendSubmission(ctx, sub)
		ids = append(ids, id)
	}

	// Get all submissions and verify
	allSubmissions, err := k.GetAllSubmissions(ctx)
	require.NoError(t, err)
	require.Len(t, allSubmissions, len(submissions))

	// Map of submissions by ID for easier comparison
	submissionMap := make(map[uint64]types.Submission)
	for i, id := range ids {
		submissionMap[id] = submissions[i]
	}

	// Verify each submission is in the returned list
	for i, got := range allSubmissions {
		expected := submissionMap[uint64(i)]
		require.Equal(t, expected.Creator, got.Creator)
		require.Equal(t, expected.Region, got.Region)
		require.Equal(t, expected.Hash, got.Hash)
		require.Equal(t, expected.Valid, got.Valid)
	}
}
