package keeper

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/ardaglobal/arda-poc/x/arda/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SubmitHash(goCtx context.Context, msg *types.MsgSubmitHash) (*types.MsgSubmitHashResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// 1. Get the pubkey for the region. Load it if it hasn't been loaded yet.
	pubKeyBase64, err := getRegionPubKey(msg.Region)
	if err != nil {
		return nil, err
	}

	pubKeyBytes, err := base64.StdEncoding.DecodeString(pubKeyBase64)
	if err != nil {
		return nil, errors.New("error decoding pubkey")
	}
	if len(pubKeyBytes) != ed25519.PublicKeySize {
		return nil, errors.New("invalid pubkey length")
	}

	// 2. Decode the signature from hex/base64 string to bytes
	sigBytes, err := hex.DecodeString(msg.Signature)
	if err != nil {
		return nil, errors.New("error decoding signature")
	}
	if len(sigBytes) != ed25519.SignatureSize {
		return nil, errors.New("invalid signature length")
	}

	// 3. Decode the hash string into bytes
	hashBytes, err := hex.DecodeString(msg.Hash)
	if err != nil {
		return nil, errors.New("error decoding hash")
	}
	// Ensure this is 32 bytes if it's a SHA-256 hex, otherwise verification will fail.
	if len(hashBytes) != 32 {
		return nil, errors.New("hash length mismatch")
	}
	// 4. Verify the signature using Ed25519
	valid := ed25519.Verify(pubKeyBytes, hashBytes, sigBytes)

	// 5. Create the Submission object with valid/invalid flag
	submission := types.Submission{
		Creator: msg.Creator, // the submitter's account address
		Region:  msg.Region,
		Hash:    msg.Hash,
		Valid:   fmt.Sprintf("%t", valid),
	}
	// 6. Append to state (store it)
	id := k.AppendSubmission(ctx, submission)

	// 7. Emit an event with details
	ctx.EventManager().EmitEvent(
		sdk.NewEvent("submission",
			sdk.NewAttribute("region", msg.Region),
			sdk.NewAttribute("hash", msg.Hash),
			sdk.NewAttribute("valid", fmt.Sprintf("%t", valid)),
		),
	)
	return &types.MsgSubmitHashResponse{Id: id}, nil
}
