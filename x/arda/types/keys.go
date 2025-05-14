package types

const (
	// ModuleName defines the module name
	ModuleName = "arda"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_arda"
)

var (
	ParamsKey = []byte("p_arda")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
