package scripts

import (
	"testing"

	"context"

	keepertest "github.com/ardaglobal/arda-poc/testutil/keeper"
	"github.com/ardaglobal/arda-poc/testutil/sample"
	"github.com/ardaglobal/arda-poc/x/arda/keeper"
	"github.com/ardaglobal/arda-poc/x/arda/types"
)

func TestGenerateHashAndSignature_SubmitHash(t *testing.T) {
	hashHex, sigHex, err := GenerateHashAndSignature()
	if err != nil {
		t.Fatalf("failed to generate hash and signature: %v", err)
	}

	_, msgServer, ctx := setupMsgServer(t)
	creator := sample.AccAddress()
	msg := types.NewMsgSubmitHash(creator, "dubai", hashHex, sigHex)
	resp, err := msgServer.SubmitHash(ctx, msg)
	if err != nil {
		t.Fatalf("SubmitHash failed: %v", err)
	}
	if resp == nil {
		t.Fatalf("SubmitHash response is nil")
	}
}

// setupMsgServer is copied from x/arda/keeper/msg_server_test.go
func setupMsgServer(t testing.TB) (keeper.Keeper, types.MsgServer, context.Context) {
	k, ctx := keepertest.ArdaKeeper(t)
	return k, keeper.NewMsgServerImpl(k), ctx
}
