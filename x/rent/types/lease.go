package types

// Lease defines a rental agreement.
type Lease struct {
	Id          string
	PropertyId  string
	Tenant      string
	RentAmount  uint64
	RentDueDate string
	Status      string
}
