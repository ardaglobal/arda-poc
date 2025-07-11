package types

import (
	"fmt"
)

// DefaultIndex is the default global index
const DefaultIndex uint64 = 1

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		LeaseList: []Lease{},
		// this line is used by starport scaffolding # genesis/types/default
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	// Check for duplicated ID in lease
	leaseIdMap := make(map[uint64]bool)
	leaseCount := gs.GetLeaseCount()
	for _, elem := range gs.LeaseList {
		if _, ok := leaseIdMap[elem.Id]; ok {
			return fmt.Errorf("duplicated id for lease")
		}
		if elem.Id >= leaseCount {
			return fmt.Errorf("lease id should be lower or equal than the last id")
		}
		leaseIdMap[elem.Id] = true
	}
	// this line is used by starport scaffolding # genesis/types/validate

	return gs.Params.Validate()
}
