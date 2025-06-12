package types

import "strings"

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

func KeyPrefix(p string) []byte {
	return []byte(p)
}

// MortgageMarkerDenom returns the sanitized bank denom for a mortgage marker token.
func MortgageMarkerDenom(collateral, index string) string {
	sanitizedCollateral := strings.ReplaceAll(collateral, " ", "")
	return "mortgage/" + sanitizedCollateral + "/" + index
}
