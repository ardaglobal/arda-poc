package types

const (
	// ModuleName defines the module name
	ModuleName = "usdarda"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_usdarda"

	// USDArdaDenom defines the bank denom for USDArda tokens
	USDArdaDenom = "usdarda"

	// MintInfoKeyPrefix stores mint info by property id
	MintInfoKeyPrefix = "MintInfo/value/"
)

var (
	ParamsKey = []byte("p_usdarda")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
