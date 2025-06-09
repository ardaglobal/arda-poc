package types

import "context"

type QueryServer interface {
	MortgageAll(context.Context, *QueryAllMortgageRequest) (*QueryAllMortgageResponse, error)
	Mortgage(context.Context, *QueryGetMortgageRequest) (*QueryGetMortgageResponse, error)
	Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error)
}

type UnimplementedQueryServer struct{}

func (*UnimplementedQueryServer) MortgageAll(context.Context, *QueryAllMortgageRequest) (*QueryAllMortgageResponse, error) {
	return nil, grpcUnimplemented
}
func (*UnimplementedQueryServer) Mortgage(context.Context, *QueryGetMortgageRequest) (*QueryGetMortgageResponse, error) {
	return nil, grpcUnimplemented
}
func (*UnimplementedQueryServer) Params(context.Context, *QueryParamsRequest) (*QueryParamsResponse, error) {
	return nil, grpcUnimplemented
}
