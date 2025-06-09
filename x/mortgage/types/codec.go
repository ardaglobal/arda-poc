package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var ModuleCdc = codec.NewProtoCodec(nil)

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateMortgage{},
		&MsgMintMortgageToken{},
		&MsgBurnMortgageToken{},
	)
	// Note: no Msg service description yet
}
