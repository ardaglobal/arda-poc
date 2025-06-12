package types

import "encoding/binary"

var _ binary.ByteOrder

const (
	// MortgageKeyPrefix is the prefix to retrieve all Mortgage
	MortgageKeyPrefix = "Mortgage/value/"
)

// MortgageKey returns the store key to retrieve a Mortgage from the index fields
func MortgageKey(
	index string,
) []byte {
	var key []byte

	indexBytes := []byte(index)
	key = append(key, indexBytes...)
	key = append(key, []byte("/")...)

	return key
}
