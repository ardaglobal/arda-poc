package keeper_test

import (
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/ardaglobal/arda-poc/testutil/keeper"
	"github.com/ardaglobal/arda-poc/testutil/nullify"
	"github.com/ardaglobal/arda-poc/x/mortgage/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestMortgageQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.MortgageKeeper(t)
	msgs := createNMortgage(keeper, ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetMortgageRequest
		response *types.QueryGetMortgageResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetMortgageRequest{
				Index: msgs[0].Index,
			},
			response: &types.QueryGetMortgageResponse{Mortgage: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetMortgageRequest{
				Index: msgs[1].Index,
			},
			response: &types.QueryGetMortgageResponse{Mortgage: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetMortgageRequest{
				Index: strconv.Itoa(100000),
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.Mortgage(ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t,
					nullify.Fill(tc.response),
					nullify.Fill(response),
				)
			}
		})
	}
}

func TestMortgageQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.MortgageKeeper(t)
	msgs := createNMortgage(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllMortgageRequest {
		return &types.QueryAllMortgageRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.MortgageAll(ctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Mortgage), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.Mortgage),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.MortgageAll(ctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Mortgage), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.Mortgage),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.MortgageAll(ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.Mortgage),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.MortgageAll(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
