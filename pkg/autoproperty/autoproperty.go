package autoproperty

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"
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

// intSliceToString converts []int to comma separated string
func intSliceToString(v []int) string {
	parts := make([]string, len(v))
	for i, s := range v {
		parts[i] = fmt.Sprintf("%d", s)
	}
	return strings.Join(parts, ",")
}

// registerProperty creates a new property with random data and sends the CLI command.
func registerProperty(users []string) Property {
	address := fmt.Sprintf("%d Main St", rand.Intn(9000)+100)
	value := rand.Intn(900000) + 100000

	// choose random number of owners
	n := rand.Intn(len(users)) + 1
	shuffledUsers := make([]string, len(users))
	copy(shuffledUsers, users)
	rand.Shuffle(len(shuffledUsers), func(i, j int) { shuffledUsers[i], shuffledUsers[j] = shuffledUsers[j], shuffledUsers[i] })
	owners := make([]string, n)
	copy(owners, shuffledUsers[:n])

	shares := randomShares(n)

	ownersStr := strings.Join(owners, ",")
	sharesStr := intSliceToString(shares)

	args := []string{
		"tx", "property", "register-property",
		address, "dubai", fmt.Sprintf("%d", value),
		"--owners", ownersStr,
		"--shares", sharesStr,
		"--gas", "auto",
		"--from", "ERES",
		"-y",
	}

	cmd := exec.Command("arda-pocd", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
		fmt.Println("Stdout:", stdout.String())
		fmt.Println("Stderr:", stderr.String())
	if err != nil {
		fmt.Println("Error registering property:", err)
		os.Exit(1)
	}

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
		strings.ToLower(p.Address),
		fromOwner, fmt.Sprintf("%d", share),
		toOwner, fmt.Sprintf("%d", share),
		"--gas", "auto",
		"--from", "ERES",
		"-y",
	}

	cmd := exec.Command("arda-pocd", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
		fmt.Println("Stdout:", stdout.String())
		fmt.Println("Stderr:", stderr.String())
	if err != nil {
		fmt.Println("Error transferring shares:", err)
		os.Exit(1)
	}

	p.Shares[fromIdx] -= share
	p.Shares[toIdx] += share
}

// Run registers properties and continuously transfers shares.
func Run(userList []string) {
	if len(userList) < 2 {
		fmt.Println("AutoProperty: At least 2 users are required to run, skipping.")
		return
	}

	rand.Seed(time.Now().UnixNano())

	var properties []Property
	for i := 0; i < 10; i++ {
		fmt.Println("AutoProperty: Registering property...")
		p := registerProperty(userList)
		properties = append(properties, p)
		time.Sleep(5 * time.Second)

		fmt.Println("AutoProperty: Creating transfer...")
		randIdx := rand.Intn(len(properties))
		transferShares(&properties[randIdx])
		time.Sleep(5 * time.Second)
	}
} 