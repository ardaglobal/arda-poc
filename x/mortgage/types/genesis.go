package types

import (
	"fmt"
)

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		MortgageList: []Mortgage{},
		// this line is used by starport scaffolding # genesis/types/default
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Check for duplicated index in mortgage
	mortgageIndexMap := make(map[string]struct{})

	for _, elem := range gs.MortgageList {
		index := string(MortgageKey(elem.Index))
		if _, ok := mortgageIndexMap[index]; ok {
			return fmt.Errorf("duplicated index for mortgage")
		}
		mortgageIndexMap[index] = struct{}{}
	}
	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
