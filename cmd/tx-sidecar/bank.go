package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	fiber "github.com/gofiber/fiber/v2"

	"cosmossdk.io/math"
	mortgagetypes "github.com/ardaglobal/arda-poc/x/mortgage/types"
	propertytypes "github.com/ardaglobal/arda-poc/x/property/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/google/uuid"
	zlog "github.com/rs/zerolog/log"
)

// MortgageRequest stores information about a user's request for a mortgage.
type MortgageRequest struct {
	ID           string    `json:"id"`
	Requester    string    `json:"requester"`   // Name of the user (lendee) who made the request.
	Lender       string    `json:"lender"`      // Name of the user (lender) the request is for.
	LendeeAddr   string    `json:"lendee_addr"` // Bech32 address of the lendee.
	Index        string    `json:"index"`
	Collateral   string    `json:"collateral"`
	Amount       uint64    `json:"amount"`
	InterestRate string    `json:"interest_rate"`
	Term         string    `json:"term"`
	Status       string    `json:"status"` // e.g., "pending", "completed"
	Timestamp    time.Time `json:"timestamp"`

	// Property purchase details
	PropertyID string   `json:"property_id,omitempty"`
	FromOwners []string `json:"from_owners,omitempty"`
	FromShares []uint64 `json:"from_shares,omitempty"`
	ToOwners   []string `json:"to_owners,omitempty"`
	ToShares   []uint64 `json:"to_shares,omitempty"`
	Price      uint64   `json:"price,omitempty"`
}

// MortgageRequestPayload is used for both requesting and creating a mortgage, including property purchase details.
type MortgageRequestPayload struct {
	Lender       string `json:"lender"`
	Index        string `json:"index"`
	Collateral   string `json:"collateral"`
	Amount       uint64 `json:"amount"`
	InterestRate string `json:"interest_rate"`
	Term         string `json:"term"`
	Gas          string `json:"gas,omitempty"`

	// Property purchase details
	PropertyID string   `json:"property_id"`
	FromOwners []string `json:"from_owners"`
	FromShares []uint64 `json:"from_shares"`
	ToOwners   []string `json:"to_owners"`
	ToShares   []uint64 `json:"to_shares"`
	Price      uint64   `json:"price"`
	Lendee     string   `json:"lendee,omitempty"` // Only used for create (lender approval)
}

// RepayMortgageRequest defines the request body for repaying a mortgage.
type RepayMortgageRequest struct {
	MortgageID string `json:"mortgage_id"`
	Amount     uint64 `json:"amount"`
	Gas        string `json:"gas,omitempty"`
}

// createMortgageHandler handles the creation of a mortgage, approving a pending request.
// @Summary Create a mortgage (lender)
// @Description Submits a transaction to create a new mortgage, effectively approving a pending request. This must be called by the **lender**, who must be logged in. The sidecar will use the logged-in user's account to sign the transaction, funding the mortgage from their account. The details in the request body should match the details from a pending mortgage request.
// @Accept json
// @Produce json
// @Param request body MortgageRequestPayload true "mortgage details (with property purchase details)"
// @Success 200 {object} map[string]string{tx_hash=string}
// @Router /bank/mortgage/create [post]
func (s *Server) createMortgageHandler(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodPost {
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Invalid request method")
	}

	if s.loggedInUser == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("No user is logged in. The lender must be logged in to create a mortgage.")
	}

	var req MortgageRequestPayload
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
	}

	zlog.Info().Str("handler", "createMortgageHandler").Interface("request", req).Msg("received request")

	fromName := s.loggedInUser
	msgBuilder := func(fromAddr string) sdk.Msg {
		return mortgagetypes.NewMsgCreateMortgage(
			fromAddr, // creator
			req.Index,
			fromAddr, // lender is the creator
			req.Lendee,
			req.Collateral,
			req.Amount,
			req.InterestRate,
			req.Term,
		)
	}

	txHash, err := s.buildSignAndBroadcastInternal(c.Context(), fromName, req.Gas, "create_mortgage", msgBuilder)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	// After successfully broadcasting, update the original request's status
	for i, mr := range s.mortgageRequests {
		if mr.Index == req.Index && mr.Lender == fromName && mr.Status == "pending" {
			s.mortgageRequests[i].Status = "completed"
			// After mortgage approval, immediately send transfer shares request if property purchase details are present
			if req.PropertyID != "" && len(req.FromOwners) > 0 && len(req.FromShares) > 0 && len(req.ToOwners) > 0 && len(req.ToShares) > 0 {
				_, err := s.buildSignAndBroadcastInternal(c.Context(), "ERES", req.Gas, "transfer_shares", func(fromAddr string) sdk.Msg {
					return propertytypes.NewMsgTransferShares(
						fromAddr,
						req.PropertyID,
						req.FromOwners,
						req.FromShares,
						req.ToOwners,
						req.ToShares,
					)
				})
				if err != nil {
					zlog.Error().Err(err).Msg("failed to broadcast transfer after mortgage")
				}
			}
			if err := s.saveMortgageRequestsToFile(); err != nil {
				zlog.Error().Err(err).Msgf("failed to update status for mortgage request index %s", req.Index)
			}
			break // Assume index is unique per lender
		}
	}

	return c.JSON(fiber.Map{"tx_hash": txHash})
}

// repayMortgageHandler handles the repayment of a mortgage.
// @Summary Repay a mortgage (lendee)
// @Description Submits a transaction to repay a portion of an outstanding mortgage. This must be called by the **lendee**, who must be logged in.
// @Accept json
// @Produce json
// @Param request body RepayMortgageRequest true "repayment details"
// @Success 200 {object} map[string]string{tx_hash=string}
// @Router /bank/mortgage/repay [post]
func (s *Server) repayMortgageHandler(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodPost {
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Invalid request method")
	}

	if s.loggedInUser == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("No user is logged in. The lendee must be logged in to repay a mortgage.")
	}

	var req RepayMortgageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
	}

	zlog.Info().Str("handler", "repayMortgageHandler").Interface("request", req).Msg("received request")

	fromName := s.loggedInUser
	msgBuilder := func(fromAddr string) sdk.Msg {
		return mortgagetypes.NewMsgRepayMortgage(
			fromAddr, // creator is the lendee
			req.MortgageID,
			req.Amount,
		)
	}

	return s.buildSignAndBroadcast(c, fromName, req.Gas, "repay_mortgage", msgBuilder)
}

// RequestFundsRequest defines the request body for requesting funds from the bank.
type RequestFundsRequest struct {
	Address string `json:"address"`
	Amount  uint64 `json:"amount"`
	Denom   string `json:"denom"`
	Gas     string `json:"gas,omitempty"`
}

// requestFundsHandler requests funds from the built-in faucet.
// @Summary Request funds from faucet
// @Description Requests funds from the built-in bank/faucet. This is only available for development and testing purposes. The bank account must be funded for this to work. On the first run, the sidecar will generate a `bank` account and print its mnemonic phrase to the console. This mnemonic must be used to send funds to the bank address before it can dispense tokens.
// @Accept json
// @Produce json
// @Param request body RequestFundsRequest true "request details"
// @Success 200 {object} map[string]string{tx_hash=string}
// @Router /bank/request-funds [post]
func (s *Server) requestFundsHandler(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodPost {
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Invalid request method")
	}

	var req RequestFundsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
	}

	zlog.Info().Str("handler", "requestFundsHandler").Interface("request", req).Msg("received request")

	if req.Amount == 0 || req.Denom == "" || req.Address == "" {
		return c.Status(fiber.StatusBadRequest).SendString("address, amount, and denom must be provided, and amount must be positive")
	}

	if _, err := sdk.AccAddressFromBech32(req.Address); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("Invalid recipient address: %v", err))
	}

	fromName := s.faucetName
	msgBuilder := func(fromAddr string) sdk.Msg {
		return &banktypes.MsgSend{
			FromAddress: fromAddr,
			ToAddress:   req.Address,
			Amount:      sdk.NewCoins(sdk.NewCoin(req.Denom, math.NewInt(int64(req.Amount)))),
		}
	}

	return s.buildSignAndBroadcast(c, fromName, req.Gas, "request_funds", msgBuilder)
}

// requestMortgageHandler allows a user to request a mortgage from a lender.
// @Summary Request a mortgage (lendee)
// @Description Allows a logged-in user (the lendee) to request a mortgage from a specified lender. This request is stored by the sidecar and does not submit a transaction. It creates a pending request that the lender can later approve.
// @Accept json
// @Produce json
// @Param request body MortgageRequestPayload true "mortgage request (with property purchase details)"
// @Success 201 {object} MortgageRequest
// @Router /bank/mortgage/request [post]
func (s *Server) requestMortgageHandler(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodPost {
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Invalid request method")
	}

	if s.loggedInUser == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("No user is logged in. The lendee must be logged in to request a mortgage.")
	}

	var req MortgageRequestPayload
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
	}

	zlog.Info().Str("handler", "requestMortgageHandler").Interface("request", req).Msg("received request")

	// Validate that the lender exists
	if _, ok := s.users[req.Lender]; !ok {
		return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("Lender '%s' not found.", req.Lender))
	}

	// Get the requester's (lendee's) address
	requesterData := s.users[s.loggedInUser]

	newReq := MortgageRequest{
		ID:           uuid.New().String(),
		Requester:    s.loggedInUser,
		Lender:       req.Lender,
		LendeeAddr:   requesterData.Address,
		Index:        req.Index,
		Collateral:   req.Collateral,
		Amount:       req.Amount,
		InterestRate: req.InterestRate,
		Term:         req.Term,
		Status:       "pending",
		Timestamp:    time.Now(),
		PropertyID:   req.PropertyID,
		FromOwners:   req.FromOwners,
		FromShares:   req.FromShares,
		ToOwners:     req.ToOwners,
		ToShares:     req.ToShares,
		Price:        req.Price,
	}

	s.mortgageRequests = append(s.mortgageRequests, newReq)
	if err := s.saveMortgageRequestsToFile(); err != nil {
		zlog.Error().Err(err).Msg("failed to save mortgage requests file")
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to save mortgage request")
	}
	return c.Status(fiber.StatusCreated).JSON(newReq)
}

// getMortgageRequestsHandler allows a logged-in user to retrieve their pending mortgage requests.
// @Summary Get pending mortgage requests
// @Description Allows a logged-in user to retrieve a list of all their pending mortgage requests, both those they have made (as the lendee) and those made to them (as the lender).
// @Produce json
// @Success 200 {array} MortgageRequest
// @Router /bank/mortgage/requests [get]
func (s *Server) getMortgageRequestsHandler(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodGet {
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Invalid request method")
	}

	zlog.Info().Str("handler", "getMortgageRequestsHandler").Str("loggedInUser", s.loggedInUser).Msg("received request")

	if s.loggedInUser == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("No user is logged in. A user must be logged in to view their requests.")
	}

	userRequests := make([]MortgageRequest, 0)
	for _, req := range s.mortgageRequests {
		// A user can see requests if they are the lender OR the requester.
		if req.Status == "pending" && (req.Lender == s.loggedInUser || req.Requester == s.loggedInUser) {
			userRequests = append(userRequests, req)
		}
	}

	return c.JSON(userRequests)
}

// requestEquityMortgageHandler allows a user to request a home equity mortgage against a property they own, routed to a lender for approval.
// @Summary Request a home equity mortgage (pending lender approval)
// @Description Allows a user to request a home equity mortgage against a property they own. The request is routed to the specified lender for approval. The index is set to 'equity'.
// @Accept json
// @Produce json
// @Param request body MortgageRequestPayload true "equity mortgage request"
// @Success 201 {object} MortgageRequest
// @Failure 400 {object} KYCErrorResponse
// @Failure 401 {object} KYCErrorResponse
// @Router /bank/mortgage/request-equity [post]
func (s *Server) requestEquityMortgageHandler(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodPost {
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Invalid request method")
	}
	if s.loggedInUser == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("No user is logged in. Please log in to request a home equity mortgage.")
	}
	var req MortgageRequestPayload
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
	}
	// Validate that the lender exists
	if _, ok := s.users[req.Lender]; !ok {
		return c.Status(fiber.StatusBadRequest).SendString("Lender not found.")
	}
	// Validate that the requester owns the property (simple check: must be in ToOwners)
	ownsProperty := false
	for _, owner := range req.ToOwners {
		if owner == s.loggedInUser {
			ownsProperty = true
			break
		}
	}
	if !ownsProperty {
		return c.Status(fiber.StatusBadRequest).SendString("You must be an owner of the property to request a home equity mortgage.")
	}
	requesterData := s.users[s.loggedInUser]
	newReq := MortgageRequest{
		ID:           uuid.New().String(),
		Requester:    s.loggedInUser,
		Lender:       req.Lender,
		LendeeAddr:   requesterData.Address,
		Index:        "equity",
		Collateral:   req.Collateral,
		Amount:       req.Amount,
		InterestRate: req.InterestRate,
		Term:         req.Term,
		Status:       "pending",
		Timestamp:    time.Now(),
		PropertyID:   req.PropertyID,
		FromOwners:   req.FromOwners,
		FromShares:   req.FromShares,
		ToOwners:     req.ToOwners,
		ToShares:     req.ToShares,
		Price:        req.Price,
	}
	s.mortgageRequests = append(s.mortgageRequests, newReq)
	if err := s.saveMortgageRequestsToFile(); err != nil {
		zlog.Error().Err(err).Msg("failed to save mortgage requests file")
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to save mortgage request")
	}
	return c.Status(fiber.StatusCreated).JSON(newReq)
}

func (s *Server) saveMortgageRequestsToFile() error {
	data, err := json.MarshalIndent(s.mortgageRequests, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal mortgage requests: %w", err)
	}
	return os.WriteFile(s.mortgageRequestsFile, data, 0644)
}
