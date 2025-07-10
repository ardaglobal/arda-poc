package keeper_test

import (
	"strconv"
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	keepertest "github.com/ardaglobal/arda-poc/testutil/keeper"
	"github.com/ardaglobal/arda-poc/x/mortgage/keeper"
	"github.com/ardaglobal/arda-poc/x/mortgage/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestMortgageMsgServerCreate(t *testing.T) {
	k, ctx := keepertest.MortgageKeeper(t)
	srv := keeper.NewMsgServerImpl(k)
	creator := "A"
	for i := 0; i < 5; i++ {
		expected := &types.MsgCreateMortgage{Creator: creator,
			Index: strconv.Itoa(i),
		}
		_, err := srv.CreateMortgage(ctx, expected)
		require.NoError(t, err)
		rst, found := k.GetMortgage(ctx,
			expected.Index,
		)
		require.True(t, found)
		require.Equal(t, expected.Creator, rst.Creator)
	}
}

func TestMortgageMsgServerUpdate(t *testing.T) {
	creator := "A"

	tests := []struct {
		desc    string
		request *types.MsgUpdateMortgage
		err     error
	}{
		{
			desc: "Completed",
			request: &types.MsgUpdateMortgage{Creator: creator,
				Index: strconv.Itoa(0),
			},
		},
		{
			desc: "Unauthorized",
			request: &types.MsgUpdateMortgage{Creator: "B",
				Index: strconv.Itoa(0),
			},
			err: sdkerrors.ErrUnauthorized,
		},
		{
			desc: "KeyNotFound",
			request: &types.MsgUpdateMortgage{Creator: creator,
				Index: strconv.Itoa(100000),
			},
			err: sdkerrors.ErrKeyNotFound,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			k, ctx := keepertest.MortgageKeeper(t)
			srv := keeper.NewMsgServerImpl(k)
			expected := &types.MsgCreateMortgage{Creator: creator,
				Index: strconv.Itoa(0),
			}
			_, err := srv.CreateMortgage(ctx, expected)
			require.NoError(t, err)

			_, err = srv.UpdateMortgage(ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				rst, found := k.GetMortgage(ctx,
					expected.Index,
				)
				require.True(t, found)
				require.Equal(t, expected.Creator, rst.Creator)
			}
		})
	}
}

func TestMortgageMsgServerDelete(t *testing.T) {
	creator := "A"

	tests := []struct {
		desc    string
		request *types.MsgDeleteMortgage
		err     error
	}{
		{
			desc: "Completed",
			request: &types.MsgDeleteMortgage{Creator: creator,
				Index: strconv.Itoa(0),
			},
		},
		{
			desc: "Unauthorized",
			request: &types.MsgDeleteMortgage{Creator: "B",
				Index: strconv.Itoa(0),
			},
			err: sdkerrors.ErrUnauthorized,
		},
		{
			desc: "KeyNotFound",
			request: &types.MsgDeleteMortgage{Creator: creator,
				Index: strconv.Itoa(100000),
			},
			err: sdkerrors.ErrKeyNotFound,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			k, ctx := keepertest.MortgageKeeper(t)
			srv := keeper.NewMsgServerImpl(k)

			_, err := srv.CreateMortgage(ctx, &types.MsgCreateMortgage{Creator: creator,
				Index: strconv.Itoa(0),
			})
			require.NoError(t, err)
			_, err = srv.DeleteMortgage(ctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				_, found := k.GetMortgage(ctx,
					tc.request.Index,
				)
				require.False(t, found)
			}
		})
	}
}
