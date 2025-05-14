package keeper

import (
	"encoding/binary"
	"errors"
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"arda/x/arda/types"
)

// ERES key for dubai region from .arda_data/config/priv_validator_key.json
var regionPubKeys = map[string]string{
	"dubai": "uzHG3r56+TWyPCsnO4q9V4VEqV2IDjIAkloboUTOAsM=",
}

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService store.KVStoreService
		logger       log.Logger

		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
	authority string,

) Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	return Keeper{
		cdc:          cdc,
		storeService: storeService,
		authority:    authority,
		logger:       logger,
	}
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// AppendSubmission saves a new submission in the store and returns its ID.
func (k Keeper) AppendSubmission(ctx sdk.Context, submission types.Submission) uint64 {
	kvStore := k.storeService.OpenKVStore(ctx)

	// Define prefixes
	submissionPrefix := types.KeyPrefix(types.KeyPrefixSubmission)
	countPrefix := types.KeyPrefix(types.KeyPrefixSubmissionCount)

	// Get current count
	countRecordKey := append(countPrefix, []byte{0}...)
	bz, err := kvStore.Get(countRecordKey)
	if err != nil {
		// If key is not found, count is 0. Otherwise, panic.
		if !errors.Is(err, sdkerrors.ErrKeyNotFound) {
			panic(fmt.Errorf("failed to get submission count: %w", err))
		}
		// bz remains nil if key not found, which is handled below
	}

	var count uint64
	if bz == nil {
		count = 0
	} else {
		count = binary.LittleEndian.Uint64(bz)
	}

	// Marshal the submission into bytes
	appendedValue := k.cdc.MustMarshal(&submission)

	// Store the submission
	submissionKey := append(submissionPrefix, types.GetSubmissionIDBytes(count)...)
	err = kvStore.Set(submissionKey, appendedValue)
	if err != nil {
		panic(fmt.Errorf("failed to set submission: %w", err))
	}

	// Update the count
	newCount := count + 1
	newCountBz := make([]byte, 8)
	binary.LittleEndian.PutUint64(newCountBz, newCount)
	err = kvStore.Set(countRecordKey, newCountBz) // Use the same countRecordKey
	if err != nil {
		panic(fmt.Errorf("failed to set new submission count: %w", err))
	}

	return count
}

func (k Keeper) GetSubmission(ctx sdk.Context, id uint64) (types.Submission, bool) {
	kvStore := k.storeService.OpenKVStore(ctx)

	// Use the same prefix as AppendSubmission
	submissionPrefix := types.KeyPrefix(types.KeyPrefixSubmission)
	submissionKey := append(submissionPrefix, types.GetSubmissionIDBytes(id)...)

	b, err := kvStore.Get(submissionKey)
	if err != nil {
		if errors.Is(err, sdkerrors.ErrKeyNotFound) {
			return types.Submission{}, false
		}
		panic(fmt.Errorf("failed to get submission with id %d: %w", id, err))
	}

	if b == nil {
		return types.Submission{}, false
	}

	var submission types.Submission
	k.cdc.MustUnmarshal(b, &submission)
	return submission, true
}

// GetAllSubmissions returns all submissions in the store
func (k Keeper) GetAllSubmissions(ctx sdk.Context) ([]types.Submission, error) {
	kvStore := k.storeService.OpenKVStore(ctx)
	submissionPrefix := types.KeyPrefix(types.KeyPrefixSubmission)

	// Get an iterator over all submission keys
	iterator, err := kvStore.Iterator(submissionPrefix, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get iterator: %w", err)
	}
	defer iterator.Close()

	submissions := []types.Submission{}

	// Iterate over all keys
	for ; iterator.Valid(); iterator.Next() {
		// Ensure the key has the proper format (prefix + submission ID)
		key := iterator.Key()
		if len(key) <= len(submissionPrefix) {
			continue
		}

		value := iterator.Value()
		var submission types.Submission
		k.cdc.MustUnmarshal(value, &submission)
		submissions = append(submissions, submission)
	}

	return submissions, nil
}
