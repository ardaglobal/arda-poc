package types

import (
	"context"
	"fmt"
)

// MsgServer defines the gRPC server interface for the mortgage module.
type MsgServer interface {
	CreateMortgage(context.Context, *MsgCreateMortgage) (*MsgCreateMortgageResponse, error)
	MintMortgageToken(context.Context, *MsgMintMortgageToken) (*MsgMintMortgageTokenResponse, error)
	BurnMortgageToken(context.Context, *MsgBurnMortgageToken) (*MsgBurnMortgageTokenResponse, error)
}

type UnimplementedMsgServer struct{}

func (*UnimplementedMsgServer) CreateMortgage(context.Context, *MsgCreateMortgage) (*MsgCreateMortgageResponse, error) {
	return nil, grpcUnimplemented
}
func (*UnimplementedMsgServer) MintMortgageToken(context.Context, *MsgMintMortgageToken) (*MsgMintMortgageTokenResponse, error) {
	return nil, grpcUnimplemented
}

func (*UnimplementedMsgServer) BurnMortgageToken(context.Context, *MsgBurnMortgageToken) (*MsgBurnMortgageTokenResponse, error) {
	return nil, grpcUnimplemented
}

var grpcUnimplemented = fmt.Errorf("gRPC methods not implemented")
