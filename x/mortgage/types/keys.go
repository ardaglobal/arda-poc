package types

const (
	// ModuleName defines the module name
	ModuleName = "mortgage"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_mortgage"
)

var (
	ParamsKey = []byte("p_mortgage")
)

const (
	// MortgageKey is the prefix to retrieve all Mortgage
	KeyPrefixMortgage = "Mortgage-value-"
	MintInfoKeyPrefix = "MintInfo-value-"
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
