package types

import "strings"

const (
	// ModuleName defines the module name
	ModuleName = "property"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// MemStoreKey defines the in-memory store key
	MemStoreKey = "mem_property"

	// KeyPrefixProperty is the prefix used to store properties by ID
	KeyPrefixProperty = "Property/value/"

	// PropertyShareDenomPrefix defines the prefix for property share denoms
	PropertyShareDenomPrefix = "prop"
)

var (
	ParamsKey = []byte("p_property")
)

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// PropertyShareDenom returns the bank denom used for the given property ID.
func PropertyShareDenom(id string) string {
	id = strings.ReplaceAll(id, " ", "")
	return PropertyShareDenomPrefix + id
}
