package types

const (
	// ModuleName defines the module name
	ModuleName = "property"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_property"

	// KeyPrefixProperty is the prefix used to store properties by ID
	KeyPrefixProperty = "Property/value/"
)

var (
	ParamsKey = []byte("p_property")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}
