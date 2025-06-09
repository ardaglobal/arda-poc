package keeper

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/ardaglobal/arda-poc/x/arda/types"
)

// Load region public keys from validator key files on demand.
// Keys are cached in memory after the first successful load.
var regionPubKeys = make(map[string]string)

// getRegionPubKey returns the public key for the provided region. If the key is
// not yet loaded, it attempts to read it from the validator key file. This
// avoids errors when the key file does not exist at application startup.
func getRegionPubKey(region string) (string, error) {
	if key, ok := regionPubKeys[region]; ok {
		return key, nil
	}

	// Currently all regions use the local validator key. Attempt to load it
	// from the default location.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	keyPath := filepath.Join(homeDir, ".arda-poc", "config", "priv_validator_key.json")
	keyData, err := os.ReadFile(keyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read validator key file: %w", err)
	}

	var keyFile struct {
		PubKey struct {
			Value string `json:"value"`
		} `json:"pub_key"`
	}
	if err := json.Unmarshal(keyData, &keyFile); err != nil {
		return "", fmt.Errorf("failed to parse validator key file: %w", err)
	}

	// Store the public key for dubai region and return it if requested.
	dubaiKey := keyFile.PubKey.Value
	regionPubKeys["dubai"] = dubaiKey
	if key, ok := regionPubKeys[region]; ok {
		return key, nil
	}

	fmt.Printf("Region %s doesn't exist, using dubai key: %s\n", region, dubaiKey)
	return dubaiKey, nil
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
