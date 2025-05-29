package types

const (
	ModuleName  = "usdarda"
	StoreKey    = ModuleName
	MemStoreKey = "mem_usdarda"

	KeyPrefixUsdArda = "UsdArda/value/"
)

var (
	ParamsKey = []byte("p_usdarda")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
