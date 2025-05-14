package types

import "encoding/binary"

const (
	// ModuleName defines the module name
	ModuleName = "arda"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_arda"

    // KeyPrefixSubmission is the prefix used to store submissions by ID.
    KeyPrefixSubmission = "Submission/value/"
    // KeyPrefixSubmissionCount stores the current submission count (for auto-increment ID).
    KeyPrefixSubmissionCount = "Submission/count/"
)

var (
	ParamsKey = []byte("p_arda")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
func GetSubmissionIDBytes(id uint64) []byte {
    bz := make([]byte, 8)
    binary.BigEndian.PutUint64(bz, id)
    return bz
}

func GetSubmissionIDFromBytes(bz []byte) uint64 {
    return binary.BigEndian.Uint64(bz)
}