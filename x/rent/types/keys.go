package types

const (
	ModuleName     = "rent"
	StoreKey       = ModuleName
	MemStoreKey    = "mem_rent"
	KeyPrefixLease = "Lease/value/"
)

var (
	ParamsKey = []byte("p_rent")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
