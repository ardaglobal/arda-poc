package types

const DefaultIndex uint64 = 1

type GenesisState struct {
	LeaseList []Lease
}

func DefaultGenesis() *GenesisState {
	return &GenesisState{LeaseList: []Lease{}}
}

func (gs GenesisState) Validate() error {
	return nil
}
