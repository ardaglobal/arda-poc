package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"cosmossdk.io/math"
	mortgagetypes "github.com/ardaglobal/arda-poc/x/mortgage/types"
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
}

// CreateMortgageByIDRequest defines the request body for creating a mortgage by ID.
type CreateMortgageByIDRequest struct {
	ID string `json:"id"`
}

// createMortgageHandler handles the creation of a mortgage, approving a pending request.
// @Summary Create a mortgage (lender)
// @Description Submits a transaction to create a new mortgage, effectively approving a pending request. This must be called by the **lender**, who must be logged in. The sidecar will use the logged-in user's account to sign the transaction, funding the mortgage from their account. The request body should only contain the ID of a pending mortgage request.
// @Accept json
// @Produce json
// @Param request body CreateMortgageByIDRequest true "mortgage request ID"
// @Success 200 {object} map[string]string{tx_hash=string}
// @Router /bank/mortgage/create [post]
func (s *Server) createMortgageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if s.loggedInUser == "" {
		http.Error(w, "No user is logged in. The lender must be logged in to create a mortgage.", http.StatusUnauthorized)
		return
	}

	var req CreateMortgageByIDRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		zlog.Error().Err(err).Msg("failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	zlog.Info().Str("handler", "createMortgageHandler").Interface("request", req).Msg("received request")

	// Find the mortgage request by ID
	var mr *MortgageRequest
	for i := range s.mortgageRequests {
		if s.mortgageRequests[i].ID == req.ID {
			mr = &s.mortgageRequests[i]
			break
		}
	}
	if mr == nil {
		http.Error(w, "Mortgage request not found", http.StatusNotFound)
		return
	}
	if mr.Status != "pending" {
		http.Error(w, "Mortgage request is not pending", http.StatusBadRequest)
		return
	}
	if mr.Lender != s.loggedInUser {
		http.Error(w, "Only the lender assigned to this request can approve it", http.StatusUnauthorized)
		return
	}

	fromName := s.loggedInUser
	msgBuilder := func(fromAddr string) sdk.Msg {
		return mortgagetypes.NewMsgCreateMortgage(
			fromAddr, // creator
			mr.Index,
			fromAddr, // lender is the creator
			mr.LendeeAddr,
			mr.Collateral,
			mr.Amount,
			mr.InterestRate,
			mr.Term,
		)
	}

	txHash, err := s.buildSignAndBroadcastInternal(r.Context(), fromName, "create_mortgage", msgBuilder)
	if err != nil {
		zlog.Error().Err(err).Msg("failed to build sign and broadcast")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// After successfully broadcasting, update the original request's status
	mr.Status = "completed"
	// After mortgage approval, immediately send transfer shares request if property purchase details are present
	if mr.Index != "equity" {
		if mr.PropertyID != "" && len(mr.FromOwners) > 0 && len(mr.FromShares) > 0 && len(mr.ToOwners) > 0 && len(mr.ToShares) > 0 {
			transferReq := TransferSharesRequest{
				PropertyID: mr.PropertyID,
				FromOwners: mr.FromOwners,
				FromShares: mr.FromShares,
				ToOwners:   mr.ToOwners,
				ToShares:   mr.ToShares,
			}
			transferReqBody, _ := json.Marshal(transferReq)
			r2 := &http.Request{Body: io.NopCloser(strings.NewReader(string(transferReqBody))), Method: http.MethodPost}
			s.transferSharesHandler(w, r2)
		}
	}
	if err := s.saveMortgageRequestsToFile(); err != nil {
		zlog.Error().Err(err).Msgf("failed to update status for mortgage request id %s", mr.ID)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"tx_hash": txHash})
}

// repayMortgageHandler handles the repayment of a mortgage.
// @Summary Repay a mortgage (lendee)
// @Description Submits a transaction to repay a portion of an outstanding mortgage. This must be called by the **lendee**, who must be logged in.
// @Accept json
// @Produce json
// @Param request body RepayMortgageRequest true "repayment details"
// @Success 200 {object} map[string]string{tx_hash=string}
// @Router /bank/mortgage/repay [post]
func (s *Server) repayMortgageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if s.loggedInUser == "" {
		http.Error(w, "No user is logged in. The lendee must be logged in to repay a mortgage.", http.StatusUnauthorized)
		return
	}

	var req RepayMortgageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
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

	s.buildSignAndBroadcast(w, r, fromName, "repay_mortgage", msgBuilder)
}

// RequestFundsRequest defines the request body for requesting funds from the bank.
type RequestFundsRequest struct {
	Address string `json:"address"`
	Amount  uint64 `json:"amount"`
	Denom   string `json:"denom"`
}

// requestFundsHandler requests funds from the built-in faucet.
// @Summary Request funds from faucet
// @Description Requests funds from the built-in bank/faucet. This is only available for development and testing purposes. The bank account must be funded for this to work. On the first run, the sidecar will generate a `bank` account and print its mnemonic phrase to the console. This mnemonic must be used to send funds to the bank address before it can dispense tokens.
// @Accept json
// @Produce json
// @Param request body RequestFundsRequest true "request details"
// @Success 200 {object} map[string]string{tx_hash=string}
// @Router /bank/request-funds [post]
func (s *Server) requestFundsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req RequestFundsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	zlog.Info().Str("handler", "requestFundsHandler").Interface("request", req).Msg("received request")

	if req.Amount == 0 || req.Denom == "" || req.Address == "" {
		http.Error(w, "address, amount, and denom must be provided, and amount must be positive", http.StatusBadRequest)
		return
	}

	if _, err := sdk.AccAddressFromBech32(req.Address); err != nil {
		http.Error(w, fmt.Sprintf("Invalid recipient address: %v", err), http.StatusBadRequest)
		return
	}

	fromName := s.faucetName
	msgBuilder := func(fromAddr string) sdk.Msg {
		return &banktypes.MsgSend{
			FromAddress: fromAddr,
			ToAddress:   req.Address,
			Amount:      sdk.NewCoins(sdk.NewCoin(req.Denom, math.NewInt(int64(req.Amount)))),
		}
	}

	s.buildSignAndBroadcast(w, r, fromName, "request_funds", msgBuilder)
}

// requestMortgageHandler allows a user to request a mortgage from a lender.
// @Summary Request a mortgage (lendee)
// @Description Allows a logged-in user (the lendee) to request a mortgage from a specified lender. This request is stored by the sidecar and does not submit a transaction. It creates a pending request that the lender can later approve.
// @Accept json
// @Produce json
// @Param request body MortgageRequestPayload true "mortgage request (with property purchase details)"
// @Success 201 {object} MortgageRequest
// @Router /bank/mortgage/request [post]
func (s *Server) requestMortgageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if s.loggedInUser == "" {
		http.Error(w, "No user is logged in. The lendee must be logged in to request a mortgage.", http.StatusUnauthorized)
		return
	}

	var req MortgageRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		zlog.Error().Err(err).Msg("failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	zlog.Info().Str("handler", "requestMortgageHandler").Interface("request", req).Msg("received request")

	// Validate that the lender exists
	if _, ok := s.users[req.Lender]; !ok {
		zlog.Error().Str("lender", req.Lender).Msg("lender not found")
		http.Error(w, fmt.Sprintf("Lender '%s' not found.", req.Lender), http.StatusBadRequest)
		return
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
		http.Error(w, "Failed to save mortgage request", http.StatusInternalServerError)
		zlog.Error().Err(err).Msg("failed to save mortgage requests file")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newReq)
}

// getMortgageRequestsHandler allows a logged-in user to retrieve their pending mortgage requests.
// @Summary Get pending mortgage requests
// @Description Allows a logged-in user to retrieve a list of all their pending mortgage requests, both those they have made (as the lendee) and those made to them (as the lender).
// @Produce json
// @Success 200 {array} MortgageRequest
// @Router /bank/mortgage/requests [get]
func (s *Server) getMortgageRequestsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	zlog.Info().Str("handler", "getMortgageRequestsHandler").Str("loggedInUser", s.loggedInUser).Msg("received request")

	if s.loggedInUser == "" {
		http.Error(w, "No user is logged in. A user must be logged in to view their requests.", http.StatusUnauthorized)
		return
	}

	userRequests := make([]MortgageRequest, 0)
	for _, req := range s.mortgageRequests {
		// A user can see requests if they are the lender OR the requester.
		if req.Status == "pending" && (req.Lender == s.loggedInUser || req.Requester == s.loggedInUser) {
			userRequests = append(userRequests, req)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userRequests)
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
func (s *Server) requestEquityMortgageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		zlog.Error().Msg("invalid request method")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	if s.loggedInUser == "" {
		zlog.Error().Msg("no user is logged in")
		http.Error(w, "No user is logged in. Please log in to request a home equity mortgage.", http.StatusUnauthorized)
		return
	}
	var req MortgageRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		zlog.Error().Err(err).Msg("failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	zlog.Info().Str("handler", "requestEquityMortgageHandler").Interface("request", req).Msg("received request")
	// Validate that the lender exists
	if _, ok := s.users[req.Lender]; !ok {
		zlog.Error().Str("lender", req.Lender).Msg("lender not found")
		http.Error(w, "Lender not found.", http.StatusBadRequest)
		return
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
		http.Error(w, "Failed to save mortgage request", http.StatusInternalServerError)
		zlog.Error().Err(err).Msg("failed to save mortgage requests file")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newReq)
}

func (s *Server) saveMortgageRequestsToFile() error {
	data, err := json.MarshalIndent(s.mortgageRequests, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal mortgage requests: %w", err)
	}
	return os.WriteFile(s.mortgageRequestsFile, data, 0644)
}
