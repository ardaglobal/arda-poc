package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/ardaglobal/arda-poc/x/rent/types"
)

func TestLeaseMsgServerCreate(t *testing.T) {
	_, srv, ctx := setupMsgServer(t)
	wctx := sdk.UnwrapSDKContext(ctx)

	creator := "A"
	for i := 0; i < 5; i++ {
		resp, err := srv.CreateLease(wctx, &types.MsgCreateLease{Creator: creator})
		require.NoError(t, err)
		require.Equal(t, i, int(resp.Id))
	}
}

func TestLeaseMsgServerUpdate(t *testing.T) {
	creator := "A"

	tests := []struct {
		desc    string
		request *types.MsgUpdateLease
		err     error
	}{
		{
			desc:    "Completed",
			request: &types.MsgUpdateLease{Creator: creator},
		},
		{
			desc:    "Unauthorized",
			request: &types.MsgUpdateLease{Creator: "B"},
			err:     sdkerrors.ErrUnauthorized,
		},
		{
			desc:    "Unauthorized",
			request: &types.MsgUpdateLease{Creator: creator, Id: 10},
			err:     sdkerrors.ErrKeyNotFound,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			_, srv, ctx := setupMsgServer(t)
			wctx := sdk.UnwrapSDKContext(ctx)

			_, err := srv.CreateLease(wctx, &types.MsgCreateLease{Creator: creator})
			require.NoError(t, err)

			_, err = srv.UpdateLease(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLeaseMsgServerDelete(t *testing.T) {
	creator := "A"

	tests := []struct {
		desc    string
		request *types.MsgDeleteLease
		err     error
	}{
		{
			desc:    "Completed",
			request: &types.MsgDeleteLease{Creator: creator},
		},
		{
			desc:    "Unauthorized",
			request: &types.MsgDeleteLease{Creator: "B"},
			err:     sdkerrors.ErrUnauthorized,
		},
		{
			desc:    "KeyNotFound",
			request: &types.MsgDeleteLease{Creator: creator, Id: 10},
			err:     sdkerrors.ErrKeyNotFound,
		},
	}
	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			_, srv, ctx := setupMsgServer(t)
			wctx := sdk.UnwrapSDKContext(ctx)

			_, err := srv.CreateLease(wctx, &types.MsgCreateLease{Creator: creator})
			require.NoError(t, err)
			_, err = srv.DeleteLease(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
