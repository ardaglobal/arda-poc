package keeper

import (
	"context"
	"encoding/json"
	"math/big"
	"strconv"

	"github.com/ardaglobal/arda-poc/x/rent/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"cosmossdk.io/errors"
	"cosmossdk.io/math"
)

func (k msgServer) PayRent(goCtx context.Context, msg *types.MsgPayRent) (*types.MsgPayRentResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Step 1: Validate lease exists and get lease details
	leaseId, err := strconv.ParseUint(msg.LeaseId, 10, 64)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidLeaseID, "invalid lease ID format: %s", msg.LeaseId)
	}

	lease, found := k.GetLease(ctx, leaseId)
	if !found {
		return nil, errors.Wrapf(types.ErrLeaseNotFound, "lease with ID %d not found", leaseId)
	}

	// Step 2: Validate sender is the tenant
	if lease.Tenant != msg.Creator {
		return nil, errors.Wrapf(types.ErrUnauthorized, "sender %s is not the tenant %s for lease %d", msg.Creator, lease.Tenant, leaseId)
	}

	// Step 3: Get transaction coins (rent amount from tx)
	senderAddr, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidAddress, "invalid sender address: %s", err)
	}

	// Get coins from the transaction context
	// Note: In a real implementation, you'd want to pass the coins as part of the message
	// For now, we'll use the lease's rent amount as the expected payment
	rentAmount := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(int64(lease.RentAmount))))

	// Step 4: Handle outstanding payments first
	outstandingPayments, err := k.parseOutstandingPayments(lease.PaymentsOutstanding)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidOutstandingPayments, "failed to parse outstanding payments: %s", err)
	}

	totalOutstanding := sdk.NewCoins()
	for _, payment := range outstandingPayments {
		totalOutstanding = totalOutstanding.Add(payment...)
	}

	// Determine payment allocation
	paymentAmount := rentAmount
	var remainingPayment sdk.Coins
	var clearedOutstanding []sdk.Coins

	if !totalOutstanding.IsZero() {
		// Apply payment to outstanding balances first
		if paymentAmount.IsAllLTE(totalOutstanding) {
			// Payment only covers part of outstanding
			clearedOutstanding = []sdk.Coins{paymentAmount}
			paymentAmount = sdk.NewCoins() // No remaining for current rent
		} else {
			// Payment covers all outstanding plus current rent
			clearedOutstanding = outstandingPayments
			remainingPayment = paymentAmount.Sub(totalOutstanding...)
			paymentAmount = remainingPayment
		}
	}

	// Step 5: Transfer rent from tenant to module account
	moduleAddr := authtypes.NewModuleAddress(types.ModuleName)
	totalPayment := rentAmount
	
	err = k.bankKeeper.SendCoins(ctx, senderAddr, moduleAddr, totalPayment)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInsufficientFunds, "failed to transfer rent payment: %s", err)
	}

	// Step 6: Fetch property details and ownership shares
	property, err := k.ValidatePropertyExists(ctx, lease.PropertyId)
	if err != nil {
		return nil, errors.Wrapf(types.ErrPropertyNotFound, "property validation failed: %s", err)
	}

	ownerShares := k.propertyKeeper.ConvertPropertyOwnersToMap(property)
	if len(ownerShares) == 0 {
		return nil, errors.Wrap(types.ErrNoPropertyOwners, "property has no owners")
	}

	// Step 7: Calculate total shares for pro-rata calculation
	totalShares := new(big.Int)
	for _, shares := range ownerShares {
		totalShares.Add(totalShares, new(big.Int).SetUint64(shares))
	}

	if totalShares.Cmp(big.NewInt(0)) == 0 {
		return nil, errors.Wrap(types.ErrInvalidShares, "total property shares is zero")
	}

	// Step 8: Calculate and disburse pro-rata rent portions
	rentAmountBig := new(big.Int).SetUint64(lease.RentAmount)
	disbursedTotal := sdk.NewCoins()

	for ownerAddr, shares := range ownerShares {
		// Calculate owner's portion using math/big for precision
		ownerSharesBig := new(big.Int).SetUint64(shares)
		ownerPortion := new(big.Int).Mul(rentAmountBig, ownerSharesBig)
		ownerPortion.Div(ownerPortion, totalShares)

		if ownerPortion.Cmp(big.NewInt(0)) > 0 {
			ownerCoins := sdk.NewCoins(sdk.NewCoin("stake", math.NewIntFromBigInt(ownerPortion)))
			
			// Transfer from module account to owner
			ownerAccAddr, err := sdk.AccAddressFromBech32(ownerAddr)
			if err != nil {
				return nil, errors.Wrapf(types.ErrInvalidAddress, "invalid owner address %s: %s", ownerAddr, err)
			}

			err = k.bankKeeper.SendCoins(ctx, moduleAddr, ownerAccAddr, ownerCoins)
			if err != nil {
				return nil, errors.Wrapf(types.ErrDisbursementFailed, "failed to disburse rent to owner %s: %s", ownerAddr, err)
			}

			disbursedTotal = disbursedTotal.Add(ownerCoins...)
		}
	}

	// Step 9: Update lease status and payment record
	lease.Status = "good"
	lease.LastPaymentBlock = uint64(ctx.BlockHeight())

	// Step 10: Update outstanding payments
	if len(clearedOutstanding) > 0 {
		// Remove cleared payments from outstanding
		newOutstanding := []sdk.Coins{}
		for _, outstanding := range outstandingPayments {
			cleared := false
			for _, cleared_payment := range clearedOutstanding {
				if outstanding.Equal(cleared_payment) {
					cleared = true
					break
				}
			}
			if !cleared {
				newOutstanding = append(newOutstanding, outstanding)
			}
		}
		
		outstandingStr, err := k.serializeOutstandingPayments(newOutstanding)
		if err != nil {
			return nil, errors.Wrapf(types.ErrSerializationFailed, "failed to serialize outstanding payments: %s", err)
		}
		lease.PaymentsOutstanding = outstandingStr
	}

	// If no outstanding payments remain, clear the field
	if len(outstandingPayments) == 0 || (len(clearedOutstanding) > 0 && len(clearedOutstanding) == len(outstandingPayments)) {
		lease.PaymentsOutstanding = ""
	}

	// Save updated lease
	k.SetLease(ctx, lease)

	// Step 11: Emit events
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			"rent_payment",
			sdk.NewAttribute("lease_id", msg.LeaseId),
			sdk.NewAttribute("tenant", lease.Tenant),
			sdk.NewAttribute("property_id", lease.PropertyId),
			sdk.NewAttribute("payment_amount", totalPayment.String()),
			sdk.NewAttribute("disbursed_amount", disbursedTotal.String()),
			sdk.NewAttribute("block_height", strconv.FormatInt(ctx.BlockHeight(), 10)),
			sdk.NewAttribute("outstanding_cleared", strconv.Itoa(len(clearedOutstanding))),
		),
	)

	// Emit individual disbursement events
	for ownerAddr, shares := range ownerShares {
		ownerSharesBig := new(big.Int).SetUint64(shares)
		ownerPortion := new(big.Int).Mul(rentAmountBig, ownerSharesBig)
		ownerPortion.Div(ownerPortion, totalShares)

		if ownerPortion.Cmp(big.NewInt(0)) > 0 {
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					"rent_disbursement",
					sdk.NewAttribute("lease_id", msg.LeaseId),
					sdk.NewAttribute("owner", ownerAddr),
					sdk.NewAttribute("shares", strconv.FormatUint(shares, 10)),
					sdk.NewAttribute("amount", math.NewIntFromBigInt(ownerPortion).String()),
					sdk.NewAttribute("total_shares", totalShares.String()),
				),
			)
		}
	}

	return &types.MsgPayRentResponse{
		Success:            true,
		AmountPaid:         totalPayment.String(),
		AmountDisbursed:    disbursedTotal.String(),
		OutstandingCleared: int32(len(clearedOutstanding)),
	}, nil
}

// Helper function to parse outstanding payments from JSON string
func (k msgServer) parseOutstandingPayments(outstandingStr string) ([]sdk.Coins, error) {
	if outstandingStr == "" {
		return []sdk.Coins{}, nil
	}

	var payments []string
	if err := json.Unmarshal([]byte(outstandingStr), &payments); err != nil {
		return nil, err
	}

	var result []sdk.Coins
	for _, paymentStr := range payments {
		coins, err := sdk.ParseCoinsNormalized(paymentStr)
		if err != nil {
			return nil, err
		}
		result = append(result, coins)
	}

	return result, nil
}

// Helper function to serialize outstanding payments to JSON string
func (k msgServer) serializeOutstandingPayments(payments []sdk.Coins) (string, error) {
	if len(payments) == 0 {
		return "", nil
	}

	var paymentStrings []string
	for _, payment := range payments {
		paymentStrings = append(paymentStrings, payment.String())
	}

	jsonBytes, err := json.Marshal(paymentStrings)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
