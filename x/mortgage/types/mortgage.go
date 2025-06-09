package types

type Mortgage struct {
	Index        string
	Lender       string
	Lendee       string
	Collateral   string
	Amount       uint64
	InterestRate string
	Term         string
}
