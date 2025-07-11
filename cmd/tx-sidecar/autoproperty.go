package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	propertytypes "github.com/ardaglobal/arda-poc/x/property/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	zlog "github.com/rs/zerolog/log"
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

// createOffPlanProperty generates an off plan property for a random developer.
func (s *Server) createOffPlanProperty(developerUsers []string) error {
	if len(developerUsers) == 0 {
		return fmt.Errorf("no developers available")
	}
	dev := developerUsers[rand.Intn(len(developerUsers))]
	addr := fmt.Sprintf("%d Future Rd", rand.Intn(9000)+100)
	value := uint64(rand.Intn(900000) + 1000000)
	prop := OffPlanProperty{
		ID:          uuid.New().String(),
		Address:     addr,
		Region:      "dubai",
		Value:       value,
		TotalShares: 100,
		Status:      "for_sale",
		Developer:   dev,
	}
	s.offPlanProperties = append(s.offPlanProperties, prop)
	return s.saveOffPlanPropertiesToFile()
}

// listPropertyForSale creates a local for-sale listing for a property.
func (s *Server) listPropertyForSale(p *Property) error {
	if len(p.Owners) == 0 {
		return fmt.Errorf("property has no owners")
	}
	idx := rand.Intn(len(p.Owners))
	owner := p.Owners[idx]
	owned := p.Shares[idx]
	if owned == 0 {
		return nil
	}
	shareAmt := rand.Intn(owned) + 1
	listing := ForSaleProperty{
		ID:         uuid.New().String(),
		PropertyID: strings.ToLower(p.Address),
		Owner:      owner,
		Shares:     []uint64{uint64(shareAmt)},
		Price:      uint64(p.Value),
		Status:     "listed",
	}
	s.forSaleProperties = append(s.forSaleProperties, listing)
	return s.saveForSalePropertiesToFile()
}

// registerProperty creates a new property with random data and sends the CLI command.
func (s *Server) registerProperty(ctx context.Context, users []string) (Property, error) {
	address := fmt.Sprintf("%d Main St", rand.Intn(9000)+100)
	value := rand.Intn(900000) + 1000000

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

	_, err := s.buildSignAndBroadcastInternal(ctx, fromName, "register_property", msgBuilder)
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

	_, err := s.buildSignAndBroadcastInternal(ctx, fromName, "transfer_shares", msgBuilder)
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

	_, err := s.buildSignAndBroadcastInternal(ctx, fromName, "edit_property_metadata", msgBuilder)
	if err != nil {
		return fmt.Errorf("autoproperty failed to edit property metadata: %w", err)
	}

	zlog.Info().Msgf("AutoProperty: Successfully edited metadata for property %s", p.Address)
	return nil
}

// RunAutoProperty registers properties and continuously transfers shares.
func (s *Server) RunAutoProperty(developerUsers, investorUsers []string) {
	if len(developerUsers) == 0 {
		zlog.Info().Msg("AutoProperty: At least 1 developer is required to run, skipping.")
		return
	}
	if len(investorUsers) < 2 {
		zlog.Info().Msg("AutoProperty: At least 2 investors are required to run, skipping.")
		return
	}

	ctx := context.Background()

	// Initial setup: create ten off plan properties for more test data
	for i := 0; i < 10; i++ {
		if err := s.createOffPlanProperty(developerUsers); err != nil {
			zlog.Error().Err(err).Msg("autoproperty create off plan")
		}
	}

	// Initial batch creation: create 20 properties quickly for testing pagination
	zlog.Info().Msg("AutoProperty: Creating initial batch of properties for testing...")
	for i := 0; i < 20; i++ {
		zlog.Info().Msgf("AutoProperty: Creating initial property %d/20...", i+1)
		
		p, err := s.registerProperty(ctx, developerUsers)
		if err != nil {
			zlog.Error().Err(err).Msgf("autoproperty batch register property %d", i+1)
			continue
		}

		// Edit metadata for the property
		if err := s.autoEditPropertyMetadata(ctx, &p); err != nil {
			zlog.Error().Err(err).Msgf("autoproperty batch edit metadata %d", i+1)
		}

		// Add some variety by randomly listing for sale or transferring shares
		if rand.Intn(3) == 0 { // 33% chance to list for sale
			if err := s.listPropertyForSale(&p); err != nil {
				zlog.Error().Err(err).Msgf("autoproperty batch list for sale %d", i+1)
			}
		}
		
		if rand.Intn(2) == 0 { // 50% chance to transfer shares
			if err := s.transferShares(ctx, &p, investorUsers); err != nil {
				zlog.Error().Err(err).Msgf("autoproperty batch transfer %d", i+1)
			}
		}

		// Short delay to avoid overwhelming the system
		time.Sleep(1 * time.Second)
	}
	
	zlog.Info().Msg("AutoProperty: Initial batch creation complete. Starting continuous mode...")

	// Continuous property creation loop
	for {
		zlog.Info().Msg("AutoProperty: Adding a new property...")

		// Register a new property
		p, err := s.registerProperty(ctx, developerUsers)
		if err != nil {
			zlog.Error().Err(err).Msg("autoproperty register property")
			time.Sleep(10 * time.Second)
			continue
		}

		// Edit property metadata
		zlog.Info().Msg("AutoProperty: Editing property metadata...")
		if err := s.autoEditPropertyMetadata(ctx, &p); err != nil {
			zlog.Error().Err(err).Msg("autoproperty edit metadata")
		}

		// Randomly decide to list for sale or transfer shares
		if rand.Intn(2) == 0 {
			if err := s.listPropertyForSale(&p); err != nil {
				zlog.Error().Err(err).Msg("autoproperty list for sale")
			}
		} else {
			if err := s.transferShares(ctx, &p, investorUsers); err != nil {
				zlog.Error().Err(err).Msg("autoproperty transfer")
			}
		}

		zlog.Info().Msg("AutoProperty: Done with this property. Waiting 3 seconds before next one...")
		time.Sleep(3 * time.Second)
	}
}
