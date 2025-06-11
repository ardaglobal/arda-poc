package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	propertytypes "github.com/ardaglobal/arda-poc/x/property/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Property tracks current owners and their shares
// Shares are represented as percentages and should sum to 100
type Property struct {
	Address string
	Value   int
	Owners  []string
	Shares  []int
}

// randomShares generates n positive share percentages that sum to 100
func randomShares(n int) []int {
	shares := make([]int, n)
	remaining := 100
	for i := 0; i < n; i++ {
		if i == n-1 {
			shares[i] = remaining
		} else {
			// ensure at least 1 share remains for others
			max := remaining - (n - i - 1)
			share := rand.Intn(max) + 1
			shares[i] = share
			remaining -= share
		}
	}
	return shares
}

// intToUint64Slice converts []int to []uint64
func intToUint64Slice(v []int) []uint64 {
	u := make([]uint64, len(v))
	for i, s := range v {
		u[i] = uint64(s)
	}
	return u
}

// registerProperty creates a new property with random data and sends the CLI command.
func (s *Server) registerProperty(ctx context.Context, users []string) (Property, error) {
	address := fmt.Sprintf("%d Main St", rand.Intn(9000)+100)
	value := rand.Intn(900000) + 100000

	// choose random number of owners
	n := rand.Intn(len(users)) + 1
	shuffledUsers := make([]string, len(users))
	copy(shuffledUsers, users)
	rand.Shuffle(len(shuffledUsers), func(i, j int) { shuffledUsers[i], shuffledUsers[j] = shuffledUsers[j], shuffledUsers[i] })
	ownerNames := make([]string, n)
	copy(ownerNames, shuffledUsers[:n])

	ownerAddresses := make([]string, n)
	for i, name := range ownerNames {
		userData, ok := s.users[name]
		if !ok {
			return Property{}, fmt.Errorf("autoproperty: user %s not found in server's user map", name)
		}
		ownerAddresses[i] = userData.Address
	}

	shares := randomShares(n)

	fromName := "ERES"
	msgBuilder := func(fromAddr string) sdk.Msg {
		return propertytypes.NewMsgRegisterProperty(
			fromAddr,
			address,
			"dubai",
			uint64(value),
			ownerAddresses,
			intToUint64Slice(shares),
		)
	}

	_, err := s.buildSignAndBroadcastInternal(ctx, fromName, "auto", "register_property", msgBuilder)
	if err != nil {
		return Property{}, fmt.Errorf("autoproperty failed to register property: %w", err)
	}

	return Property{Address: address, Value: value, Owners: ownerNames, Shares: shares}, nil
}

// distributeShares generates n positive share integers that sum to a total.
// It assumes total >= n.
func distributeShares(n, total int) []int {
	shares := make([]int, n)
	// Give 1 to each to start with
	for i := range shares {
		shares[i] = 1
	}
	remaining := total - n

	// Distribute the remainder randomly
	for i := 0; i < remaining; i++ {
		shares[rand.Intn(n)]++
	}

	return shares
}

// transferShares performs a random share transfer from a current owner to a set of investors.
func (s *Server) transferShares(ctx context.Context, p *Property, investorUsers []string) error {
	if len(p.Owners) == 0 {
		return fmt.Errorf("autoproperty: property has no owners to transfer from")
	}
	if len(investorUsers) == 0 {
		return fmt.Errorf("autoproperty: no investors to transfer to")
	}

	// 1. Select a random "from" owner from the current property owners
	fromIdx := rand.Intn(len(p.Owners))
	fromOwnerName := p.Owners[fromIdx]
	maxShares := p.Shares[fromIdx]

	if maxShares == 0 {
		return nil // Skip transfer if the selected owner has no shares
	}

	// 2. Determine how many shares to transfer (between 1 and all of the owner's shares)
	sharesToTransfer := rand.Intn(maxShares) + 1

	// 3. Select a random number of "to" investors.
	// The number of investors cannot exceed the number of shares being transferred,
	// as each must receive at least one share.
	maxInvestors := len(investorUsers)
	if sharesToTransfer < maxInvestors {
		maxInvestors = sharesToTransfer
	}
	numToInvestors := rand.Intn(maxInvestors) + 1

	shuffledInvestors := make([]string, len(investorUsers))
	copy(shuffledInvestors, investorUsers)
	rand.Shuffle(len(shuffledInvestors), func(i, j int) {
		shuffledInvestors[i], shuffledInvestors[j] = shuffledInvestors[j], shuffledInvestors[i]
	})
	toInvestorNames := shuffledInvestors[:numToInvestors]

	// 4. Get bech32 addresses for all parties
	fromOwnerData, ok := s.users[fromOwnerName]
	if !ok {
		return fmt.Errorf("autoproperty: fromOwner '%s' not found", fromOwnerName)
	}
	fromOwnerAddress := fromOwnerData.Address

	toInvestorAddresses := make([]string, numToInvestors)
	for i, name := range toInvestorNames {
		investorData, ok := s.users[name]
		if !ok {
			return fmt.Errorf("autoproperty: toInvestor '%s' not found", name)
		}
		toInvestorAddresses[i] = investorData.Address
	}

	// 5. Distribute the shares to be transferred among the "to" investors
	toShares := distributeShares(numToInvestors, sharesToTransfer)
	toSharesUint64 := make([]uint64, len(toShares))
	for i, s := range toShares {
		toSharesUint64[i] = uint64(s)
	}

	// 6. Build and broadcast the transaction
	fromName := "ERES" // The transaction is authorized by the regulator
	msgBuilder := func(fromAddr string) sdk.Msg {
		return propertytypes.NewMsgTransferShares(
			fromAddr,
			strings.ToLower(p.Address),
			[]string{fromOwnerAddress},
			[]uint64{uint64(sharesToTransfer)},
			toInvestorAddresses,
			toSharesUint64,
		)
	}

	_, err := s.buildSignAndBroadcastInternal(ctx, fromName, "auto", "transfer_shares", msgBuilder)
	if err != nil {
		return fmt.Errorf("autoproperty failed to transfer shares: %w", err)
	}

	// 7. Update the local property state
	// First, build a map of the current state
	ownerShareMap := make(map[string]int)
	for i, owner := range p.Owners {
		ownerShareMap[owner] = p.Shares[i]
	}

	// Decrease from-owner shares
	ownerShareMap[fromOwnerName] -= sharesToTransfer

	// Update/add to-investor shares
	for i, investorName := range toInvestorNames {
		ownerShareMap[investorName] += toShares[i]
	}

	// Rebuild the owners and shares slices from the map, removing owners with 0 shares
	newOwners := make([]string, 0, len(ownerShareMap))
	newShares := make([]int, 0, len(ownerShareMap))
	for owner, share := range ownerShareMap {
		if share > 0 {
			newOwners = append(newOwners, owner)
			newShares = append(newShares, share)
		}
	}
	p.Owners = newOwners
	p.Shares = newShares

	return nil
}

// autoEditPropertyMetadata fills in placeholder data for a newly created property.
func (s *Server) autoEditPropertyMetadata(ctx context.Context, p *Property) error {
	if len(p.Owners) == 0 {
		return fmt.Errorf("autoproperty: property has no owners to source owner info from")
	}

	// 1. Generate placeholder data
	propertyName := fmt.Sprintf("Auto Property %d", rand.Intn(1000))
	propertyType := "residential"
	parcelNumber := fmt.Sprintf("PN-%d", rand.Intn(100000))
	size := fmt.Sprintf("%d sqft", rand.Intn(3000)+500)
	constructionInfo := fmt.Sprintf("Built in %d", rand.Intn(50)+1970)
	zoning := "R-1"
	// Use the first owner's name for owner information.
	// In a real scenario, this would be more detailed.
	ownerInfo := p.Owners[0]
	tenantId := "" // No tenant initially
	unitNumber := fmt.Sprintf("Unit %d", rand.Intn(100)+1)

	// 2. Build and broadcast the transaction
	fromName := "ERES" // The transaction is authorized by the regulator
	msgBuilder := func(fromAddr string) sdk.Msg {
		return propertytypes.NewMsgEditPropertyMetadata(
			fromAddr,
			strings.ToLower(p.Address), // propertyID is the address
			propertyName,
			propertyType,
			parcelNumber,
			size,
			constructionInfo,
			zoning,
			ownerInfo,
			tenantId,
			unitNumber,
		)
	}

	_, err := s.buildSignAndBroadcastInternal(ctx, fromName, "auto", "edit_property_metadata", msgBuilder)
	if err != nil {
		return fmt.Errorf("autoproperty failed to edit property metadata: %w", err)
	}

	fmt.Printf("AutoProperty: Successfully edited metadata for property %s\n", p.Address)
	return nil
}

// RunAutoProperty registers properties and continuously transfers shares.
func (s *Server) RunAutoProperty(developerUsers, investorUsers []string) {
	if len(developerUsers) == 0 {
		fmt.Println("AutoProperty: At least 1 developer is required to run, skipping.")
		return
	}
	if len(investorUsers) < 2 {
		fmt.Println("AutoProperty: At least 2 investors are required to run, skipping.")
		return
	}

	var properties []Property
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		fmt.Println("AutoProperty: Registering property...")
		p, err := s.registerProperty(ctx, developerUsers)
		if err != nil {
			fmt.Println(err)
		} else {
			properties = append(properties, p)

			// Also edit the property's metadata with placeholder info.
			fmt.Println("AutoProperty: Editing property metadata...")
			if err := s.autoEditPropertyMetadata(ctx, &properties[len(properties)-1]); err != nil {
				fmt.Println(err)
			}
		}

		if len(properties) > 0 {
			fmt.Println("AutoProperty: Creating transfer...")
			randIdx := rand.Intn(len(properties))
			err := s.transferShares(ctx, &properties[randIdx], investorUsers)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	fmt.Println("AutoProperty: Done")
} 