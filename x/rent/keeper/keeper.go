package keeper

import (
	"fmt"

	"cosmossdk.io/core/store"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ardaglobal/arda-poc/x/rent/types"
)

type (
	Keeper struct {
		cdc          codec.BinaryCodec
		storeService store.KVStoreService
		logger       log.Logger

		// the address capable of executing a MsgUpdateParams message. Typically, this
		// should be the x/gov module account.
		authority string

		propertyKeeper types.PropertyKeeper
		bankKeeper     types.BankKeeper
		accountKeeper  types.AccountKeeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeService store.KVStoreService,
	logger log.Logger,
	authority string,

	propertyKeeper types.PropertyKeeper,
	bankKeeper types.BankKeeper,
	accountKeeper types.AccountKeeper,
) Keeper {
	if _, err := sdk.AccAddressFromBech32(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	return Keeper{
		cdc:          cdc,
		storeService: storeService,
		authority:    authority,
		logger:       logger,

		propertyKeeper: propertyKeeper,
		bankKeeper:     bankKeeper,
		accountKeeper:  accountKeeper,
	}
}

// GetAuthority returns the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}

// Logger returns a module-specific logger.
func (k Keeper) Logger() log.Logger {
	return k.logger.With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetPropertyKeeper returns the property keeper.
func (k Keeper) GetPropertyKeeper() types.PropertyKeeper {
	return k.propertyKeeper
}

// GetBankKeeper returns the bank keeper.
func (k Keeper) GetBankKeeper() types.BankKeeper {
	return k.bankKeeper
}

// GetAccountKeeper returns the account keeper.
func (k Keeper) GetAccountKeeper() types.AccountKeeper {
	return k.accountKeeper
}
