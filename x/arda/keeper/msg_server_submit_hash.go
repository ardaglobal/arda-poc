package keeper

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"

	"arda/x/arda/types"

	"github.com/btcsuite/btcd/btcec/v2"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) SubmitHash(goCtx context.Context, msg *types.MsgSubmitHash) (*types.MsgSubmitHashResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// 1. Get the pubkey for the region
	pubKeyHex, ok := regionPubKeys[msg.Region]
	if !ok {
		return nil, errors.New("region not recognized")
	}
	pubKeyBytes, err := hex.DecodeString(pubKeyHex)
	if err != nil {
		return nil, errors.New("error decoding pubkey")
	}
	
	// 2. Decode the signature from hex/base64 string to bytes
	sigBytes, err := hex.DecodeString(msg.Signature)
	// If the CLI provided signature in base64, use base64 decoding instead.
	if err != nil {
		// handle error decoding signature
		return nil, errors.New("error decoding signature")
	}
	
	// 3. Parse signature bytes into r and s values (assuming 64-byte R||S format)
	if len(sigBytes) != 64 {
		// signature length mismatch, mark as invalid
		return nil, errors.New("signature length mismatch")
	}
	r := new(big.Int).SetBytes(sigBytes[:32])
	s := new(big.Int).SetBytes(sigBytes[32:])
	
	// 4. Parse the public key bytes (33-byte compressed) into an ECDSA public key
	pubKey, err := btcec.ParsePubKey(pubKeyBytes)
	if err != nil {
		// handle error (invalid pubkey bytes)
		return nil, errors.New("invalid pubkey bytes")
	}
	ecdsaPubKey := pubKey.ToECDSA()
	
	// 5. Decode the hash string into bytes
	hashBytes, err := hex.DecodeString(msg.Hash)
	if err != nil {
		return nil, errors.New("error decoding hash")
	}
	// Ensure this is 32 bytes if it's a SHA-256 hex, otherwise verification will fail.
	if len(hashBytes) != 32 {
		return nil, errors.New("hash length mismatch")
	}
	// 6. Verify the signature using ECDSA
	// TODO: see how the UX differs if we reject vs just mark as invalid
	valid := ecdsa.Verify(ecdsaPubKey, hashBytes, r, s)
	
	// 7. Create the Submission object with valid/invalid flag
	submission := types.Submission{
		Creator: msg.Creator,       // the submitter's account address
		Region:  msg.Region,
		Hash:    msg.Hash,
		Valid:   fmt.Sprintf("%t", valid),
	}
	// 8. Append to state (store it)
	id := k.AppendSubmission(ctx, submission)
	
	// 9. Emit an event with details
	ctx.EventManager().EmitEvent(
		sdk.NewEvent("submission",
			sdk.NewAttribute("region", msg.Region),
			sdk.NewAttribute("hash", msg.Hash),
			sdk.NewAttribute("valid", fmt.Sprintf("%t", valid)),
		),
	)
	return &types.MsgSubmitHashResponse{Id: id}, nil
}
