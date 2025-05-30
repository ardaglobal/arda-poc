package types

// MintInfo stores minted and burned amounts for a property
// If Minted > Burned, the property is considered locked
// Outstanding minted amount is Minted - Burned

type MintInfo struct {
	PropertyId string
	Minted     uint64
	Burned     uint64
}
