package types

const (
	// ModuleName defines the module name
	ModuleName = "usdarda"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_usdarda"
)

var (
	ParamsKey = []byte("p_usdarda")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
