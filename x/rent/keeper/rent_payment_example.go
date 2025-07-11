package keeper

import (
	"context"
	"fmt"
	"strconv"

	"github.com/ardaglobal/arda-poc/x/rent/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"cosmossdk.io/math"
)

// Example demonstrates how the PayRent functionality works
// This is for documentation purposes only and should not be used in production
func (k Keeper) ExamplePayRentUsage(ctx sdk.Context) {
	// This is a demonstration of how the PayRent handler works
	// In a real scenario, this would be called through transactions

	// Example 1: Basic rent payment
	// Assume we have a lease with ID 1, property "prop123", tenant "tenant_addr"
	// and rent amount 1000 tokens

	// Create example lease (in practice, this would already exist)
	lease := types.Lease{
		Id:                      1,
		PropertyId:              "prop123",
		Tenant:                  "tenant_addr",
		RentAmount:              1000,
		Status:                  "active",
		PaymentsOutstanding:     "",
		LastPaymentBlock:        0,
		RecurringStatus:         true,
		CancellationPending:     false,
		CancellationInitiator:   "",
		CancellationDeadline:    0,
		RentDueDate:             "2024-01-01",
		TimePeriod:              30, // 30 days
		TermLength:              "12 months",
		PaymentTerms:            "Monthly",
		CancellationRequirements: "30 days notice",
		Creator:                 "property_owner",
	}

	// Store the lease
	k.SetLease(ctx, lease)

	// Example payment message
	msg := &types.MsgPayRent{
		Creator: "tenant_addr",
		LeaseId: "1",
	}

	// The PayRent handler would:
	// 1. Validate the lease exists and the sender is the tenant
	// 2. Check for outstanding payments
	// 3. Transfer rent from tenant to module account
	// 4. Fetch property ownership details
	// 5. Calculate pro-rata portions for each owner
	// 6. Disburse rent to property owners
	// 7. Update lease status and payment record
	// 8. Emit events

	fmt.Printf("Example PayRent message: %+v\n", msg)
	fmt.Printf("This would process rent payment for lease %s\n", msg.LeaseId)
}

// ExampleOutstandingPayments demonstrates how outstanding payments are handled
func (k Keeper) ExampleOutstandingPayments(ctx sdk.Context) {
	// Example lease with outstanding payments
	lease := types.Lease{
		Id:                  2,
		PropertyId:          "prop456",
		Tenant:              "tenant_addr2",
		RentAmount:          1500,
		Status:              "outstanding",
		PaymentsOutstanding: `["1500stake", "1500stake"]`, // 2 months outstanding
		LastPaymentBlock:    100,
	}

	k.SetLease(ctx, lease)

	// When tenant pays, the payment logic:
	// 1. Calculates total outstanding: 3000 tokens (2 x 1500)
	// 2. If payment is 1500, applies to first outstanding payment
	// 3. If payment is 3000+, clears all outstanding and applies to current
	// 4. Updates PaymentsOutstanding field accordingly

	fmt.Printf("Example lease with outstanding payments: %+v\n", lease)
}

// ExampleProRataCalculation demonstrates the pro-rata rent distribution
func (k Keeper) ExampleProRataCalculation(ctx sdk.Context) {
	// Example property with multiple owners
	// Owner A: 60% (600 shares out of 1000)
	// Owner B: 30% (300 shares out of 1000)
	// Owner C: 10% (100 shares out of 1000)
	// Total rent: 1000 tokens

	rentAmount := uint64(1000)
	ownerShares := map[string]uint64{
		"owner_a": 600,
		"owner_b": 300,
		"owner_c": 100,
	}
	totalShares := uint64(1000)

	fmt.Printf("Pro-rata calculation example:\n")
	fmt.Printf("Total rent: %d tokens\n", rentAmount)
	fmt.Printf("Total shares: %d\n", totalShares)

	for owner, shares := range ownerShares {
		// Calculate using the same logic as PayRent handler
		portion := (rentAmount * shares) / totalShares
		percentage := (shares * 100) / totalShares
		
		fmt.Printf("Owner %s: %d shares (%d%%) = %d tokens\n", 
			owner, shares, percentage, portion)
	}
}

// ExampleEventEmission shows what events are emitted during rent payment
func (k Keeper) ExampleEventEmission(ctx sdk.Context) {
	// Events emitted during PayRent:
	
	// 1. Main rent payment event
	event1 := sdk.NewEvent(
		"rent_payment",
		sdk.NewAttribute("lease_id", "1"),
		sdk.NewAttribute("tenant", "tenant_addr"),
		sdk.NewAttribute("property_id", "prop123"),
		sdk.NewAttribute("payment_amount", "1000stake"),
		sdk.NewAttribute("disbursed_amount", "1000stake"),
		sdk.NewAttribute("block_height", strconv.FormatInt(ctx.BlockHeight(), 10)),
		sdk.NewAttribute("outstanding_cleared", "0"),
	)

	// 2. Individual disbursement events (one per owner)
	event2 := sdk.NewEvent(
		"rent_disbursement",
		sdk.NewAttribute("lease_id", "1"),
		sdk.NewAttribute("owner", "owner_a"),
		sdk.NewAttribute("shares", "600"),
		sdk.NewAttribute("amount", "600"),
		sdk.NewAttribute("total_shares", "1000"),
	)

	fmt.Printf("Example events:\n")
	fmt.Printf("Payment event: %+v\n", event1)
	fmt.Printf("Disbursement event: %+v\n", event2)
}

// ExampleErrorHandling shows common error scenarios
func (k Keeper) ExampleErrorHandling() {
	fmt.Printf("Common error scenarios in PayRent:\n")
	fmt.Printf("1. ErrLeaseNotFound - lease ID doesn't exist\n")
	fmt.Printf("2. ErrUnauthorized - sender is not the tenant\n")
	fmt.Printf("3. ErrInsufficientFunds - tenant doesn't have enough tokens\n")
	fmt.Printf("4. ErrPropertyNotFound - property doesn't exist\n")
	fmt.Printf("5. ErrNoPropertyOwners - property has no owners\n")
	fmt.Printf("6. ErrDisbursementFailed - failed to send tokens to owner\n")
	fmt.Printf("7. ErrInvalidOutstandingPayments - malformed outstanding payments JSON\n")
}