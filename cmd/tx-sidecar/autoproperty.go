package main

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

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

// transferShares performs a random share transfer between owners of the property.
func (s *Server) transferShares(ctx context.Context, p *Property) error {
	if len(p.Owners) < 2 {
		return nil
	}

	fromIdx := rand.Intn(len(p.Owners))
	toIdx := rand.Intn(len(p.Owners))
	for toIdx == fromIdx {
		toIdx = rand.Intn(len(p.Owners))
	}

	maxShare := p.Shares[fromIdx]
	if maxShare == 0 {
		return nil
	}
	share := rand.Intn(maxShare) + 1

	fromOwnerName := p.Owners[fromIdx]
	toOwnerName := p.Owners[toIdx]

	fromOwnerData, ok := s.users[fromOwnerName]
	if !ok {
		return fmt.Errorf("autoproperty: fromOwner %s not found in server's user map", fromOwnerName)
	}
	toOwnerData, ok := s.users[toOwnerName]
	if !ok {
		return fmt.Errorf("autoproperty: toOwner %s not found in server's user map", toOwnerName)
	}
	fromOwnerAddress := fromOwnerData.Address
	toOwnerAddress := toOwnerData.Address

	fromName := "ERES"
	msgBuilder := func(fromAddr string) sdk.Msg {
		return propertytypes.NewMsgTransferShares(
			fromAddr,
			strings.ToLower(p.Address),
			[]string{fromOwnerAddress},
			[]uint64{uint64(share)},
			[]string{toOwnerAddress},
			[]uint64{uint64(share)},
		)
	}
	_, err := s.buildSignAndBroadcastInternal(ctx, fromName, "auto", "transfer_shares", msgBuilder)
	if err != nil {
		return fmt.Errorf("autoproperty failed to transfer shares: %w", err)
	}

	p.Shares[fromIdx] -= share
	p.Shares[toIdx] += share
	return nil
}

// RunAutoProperty registers properties and continuously transfers shares.
func (s *Server) RunAutoProperty(userList []string) {
	if len(userList) < 2 {
		fmt.Println("AutoProperty: At least 2 users are required to run, skipping.")
		return
	}
	rand.Seed(time.Now().UnixNano())

	var properties []Property
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		fmt.Println("AutoProperty: Registering property...")
		p, err := s.registerProperty(ctx, userList)
		if err != nil {
			fmt.Println(err)
		} else {
			properties = append(properties, p)
		}

		if len(properties) > 0 {
			fmt.Println("AutoProperty: Creating transfer...")
			randIdx := rand.Intn(len(properties))
			err := s.transferShares(ctx, &properties[randIdx])
			if err != nil {
				fmt.Println(err)
			}
		}
	}
} 