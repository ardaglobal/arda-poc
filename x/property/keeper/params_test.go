package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "arda/testutil/keeper"
	"arda/x/property/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := keepertest.PropertyKeeper(t)
	params := types.DefaultParams()

	require.NoError(t, k.SetParams(ctx, params))
	require.EqualValues(t, params, k.GetParams(ctx))
}
