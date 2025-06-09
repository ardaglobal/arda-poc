package types

import (
	"github.com/cosmos/cosmos-sdk/types/query"
)

type QueryAllMortgageRequest struct {
	Pagination *query.PageRequest
}

type QueryAllMortgageResponse struct {
	Mortgages  []*Mortgage
	Pagination *query.PageResponse
}

type QueryGetMortgageRequest struct {
	Index string
}

type QueryGetMortgageResponse struct {
	Mortgage *Mortgage
}

type QueryParamsRequest struct{}

type QueryParamsResponse struct {
	Params Params
}

// Module is used for wiring via depinject
type Module struct {
	Authority string
}
