package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "arda/testutil/keeper"
	"arda/x/arda/types"
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
} 