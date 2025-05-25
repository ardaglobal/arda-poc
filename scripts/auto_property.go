package main

import (
	"fmt"
	"math/rand"
	"os/exec"
	"strings"
	"time"
)

var users = []string{"alice", "bob", "charlie", "dan", "eve", "fred", "george", "harry"}

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

// intSliceToString converts []int to comma separated string
func intSliceToString(v []int) string {
	parts := make([]string, len(v))
	for i, s := range v {
		parts[i] = fmt.Sprintf("%d", s)
	}
	return strings.Join(parts, ",")
}

// registerProperty creates a new property with random data and sends the CLI command.
func registerProperty() Property {
	address := fmt.Sprintf("%d Main St", rand.Intn(9000)+100)
	value := rand.Intn(900000) + 100000

	// choose random number of owners
	n := rand.Intn(len(users)) + 1
	rand.Shuffle(len(users), func(i, j int) { users[i], users[j] = users[j], users[i] })
	owners := make([]string, n)
	copy(owners, users[:n])

	shares := randomShares(n)

	ownersStr := strings.Join(owners, ",")
	sharesStr := intSliceToString(shares)

	args := []string{
		"tx", "property", "register-property",
		address, "dubai", fmt.Sprintf("%d", value),
		"--owners", ownersStr,
		"--shares", sharesStr,
		"--from", "ERES",
		"-y",
	}

	cmd := exec.Command("arda-pocd", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	_ = cmd.Run() // ignore errors in this demo

	return Property{Address: address, Value: value, Owners: owners, Shares: shares}
}

// transferShares performs a random share transfer between owners of the property.
func transferShares(p *Property) {
	if len(p.Owners) < 2 {
		return
	}

	fromIdx := rand.Intn(len(p.Owners))
	toIdx := rand.Intn(len(p.Owners))
	for toIdx == fromIdx {
		toIdx = rand.Intn(len(p.Owners))
	}

	maxShare := p.Shares[fromIdx]
	if maxShare == 0 {
		return
	}
	share := rand.Intn(maxShare) + 1

	fromOwner := p.Owners[fromIdx]
	toOwner := p.Owners[toIdx]

	args := []string{
		"tx", "property", "transfer-shares",
		p.Address,
		fromOwner, fmt.Sprintf("%d", share),
		toOwner, fmt.Sprintf("%d", share),
		"--from", "ERES",
		"-y",
	}

	cmd := exec.Command("arda-pocd", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	_ = cmd.Run()

	p.Shares[fromIdx] -= share
	p.Shares[toIdx] += share
}

// AutoProperty registers properties and continuously transfers shares.
func main() {
	rand.Seed(time.Now().UnixNano())
	for {
		fmt.Println("Registering property..")
		p := registerProperty()
		for i := 0; i < 5; i++ {
			fmt.Println("  Creating transfer..")
			transferShares(&p)
			time.Sleep(1 * time.Second)
		}
		time.Sleep(2 * time.Second)
	}
}
