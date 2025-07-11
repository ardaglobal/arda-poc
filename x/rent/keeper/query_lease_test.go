package keeper_test

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/ardaglobal/arda-poc/testutil/keeper"
	"github.com/ardaglobal/arda-poc/testutil/nullify"
	"github.com/ardaglobal/arda-poc/x/rent/types"
)

func TestLeaseQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.RentKeeper(t)
	msgs := createNLease(keeper, ctx, 2)
	tests := []struct {
		desc     string
		request  *types.QueryGetLeaseRequest
		response *types.QueryGetLeaseResponse
		err      error
	}{
		{
			desc:     "First",
			request:  &types.QueryGetLeaseRequest{Id: msgs[0].Id},
			response: &types.QueryGetLeaseResponse{Lease: msgs[0]},
		},
		{
			desc:     "Second",
			request:  &types.QueryGetLeaseRequest{Id: msgs[1].Id},
			response: &types.QueryGetLeaseResponse{Lease: msgs[1]},
		},
		{
			desc:    "KeyNotFound",
			request: &types.QueryGetLeaseRequest{Id: uint64(len(msgs))},
			err:     sdkerrors.ErrKeyNotFound,
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.Lease(ctx, tc.request)
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

func TestLeaseQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.RentKeeper(t)
	msgs := createNLease(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllLeaseRequest {
		return &types.QueryAllLeaseRequest{
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
			resp, err := keeper.LeaseAll(ctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Lease), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.Lease),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.LeaseAll(ctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.Lease), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.Lease),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.LeaseAll(ctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.Lease),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.LeaseAll(ctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
