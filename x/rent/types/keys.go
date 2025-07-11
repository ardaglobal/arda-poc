package types

const (
	// ModuleName defines the module name
	ModuleName = "rent"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_rent"
)

var (
	ParamsKey = []byte("p_rent")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

const (
	LeaseKey      = "Lease/value/"
	LeaseCountKey = "Lease/count/"
)
