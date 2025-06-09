package types

const (
	// ModuleName defines the module name
	ModuleName = "mortgage"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_mortgage"

	// KeyPrefixMortgage is the prefix used to store mortgages
	KeyPrefixMortgage = "Mortgage/value/"

	// RouterKey is the message route key for the mortgage module
	RouterKey = ModuleName
)

var (
	ParamsKey = []byte("p_mortgage")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
