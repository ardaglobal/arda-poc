package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	Requester    string    `json:"requester"`     // Name of the user (lendee) who made the request.
	Lender       string    `json:"lender"`        // Name of the user (lender) the request is for.
	LendeeAddr   string    `json:"lendee_addr"`   // Bech32 address of the lendee.
	Index        string    `json:"index"`
	Collateral   string    `json:"collateral"`
	Amount       uint64    `json:"amount"`
	InterestRate string    `json:"interest_rate"`
	Term         string    `json:"term"`
	Status       string    `json:"status"` // e.g., "pending", "completed"
	Timestamp    time.Time `json:"timestamp"`
}

// NewMortgageRequest defines the request body for requesting a mortgage.
type NewMortgageRequest struct {
	Lender       string `json:"lender"`
	Index        string `json:"index"`
	Collateral   string `json:"collateral"`
	Amount       uint64 `json:"amount"`
	InterestRate string `json:"interest_rate"`
	Term         string `json:"term"`
}

// CreateMortgageRequest defines the request body for creating a mortgage.
type CreateMortgageRequest struct {
	Index        string `json:"index"`
	Lendee       string `json:"lendee"`
	Collateral   string `json:"collateral"`
	Amount       uint64 `json:"amount"`
	InterestRate string `json:"interest_rate"`
	Term         string `json:"term"`
	Gas          string `json:"gas,omitempty"`
}

// RepayMortgageRequest defines the request body for repaying a mortgage.
type RepayMortgageRequest struct {
	MortgageID string `json:"mortgage_id"`
	Amount     uint64 `json:"amount"`
	Gas        string `json:"gas,omitempty"`
}

func (s *Server) createMortgageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if s.loggedInUser == "" {
		http.Error(w, "No user is logged in. The lender must be logged in to create a mortgage.", http.StatusUnauthorized)
		return
	}

	var req CreateMortgageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
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

	txHash, err := s.buildSignAndBroadcastInternal(r.Context(), fromName, req.Gas, "create_mortgage", msgBuilder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// After successfully broadcasting, update the original request's status
	for i, mr := range s.mortgageRequests {
		if mr.Index == req.Index && mr.Lender == fromName && mr.Status == "pending" {
			s.mortgageRequests[i].Status = "completed"
			if err := s.saveMortgageRequestsToFile(); err != nil {
				// Log the error, but don't fail the whole request since the tx is already broadcast.
				zlog.Error().Err(err).Msgf("failed to update status for mortgage request index %s", req.Index)
			}
			break // Assume index is unique per lender
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"tx_hash": txHash})
}

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

	s.buildSignAndBroadcast(w, r, fromName, req.Gas, "repay_mortgage", msgBuilder)
}

// RequestFundsRequest defines the request body for requesting funds from the bank.
type RequestFundsRequest struct {
	Address string `json:"address"`
	Amount  uint64 `json:"amount"`
	Denom   string `json:"denom"`
	Gas     string `json:"gas,omitempty"`
}

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

	s.buildSignAndBroadcast(w, r, fromName, req.Gas, "request_funds", msgBuilder)
}

func (s *Server) requestMortgageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if s.loggedInUser == "" {
		http.Error(w, "No user is logged in. The lendee must be logged in to request a mortgage.", http.StatusUnauthorized)
		return
	}

	var req NewMortgageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	zlog.Info().Str("handler", "requestMortgageHandler").Interface("request", req).Msg("received request")

	// Validate that the lender exists
	if _, ok := s.users[req.Lender]; !ok {
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

func (s *Server) getMortgageRequestsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	zlog.Info().Str("handler", "getMortgageRequestsHandler").Str("loggedInUser", s.loggedInUser).Msg("received request")

	if s.loggedInUser == "" {
		http.Error(w, "No user is logged in. A lender must be logged in to view their requests.", http.StatusUnauthorized)
		return
	}

	var requestsForLender []MortgageRequest
	for _, req := range s.mortgageRequests {
		if req.Lender == s.loggedInUser && req.Status == "pending" {
			requestsForLender = append(requestsForLender, req)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requestsForLender)
}

func (s *Server) saveMortgageRequestsToFile() error {
	data, err := json.MarshalIndent(s.mortgageRequests, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal mortgage requests: %w", err)
	}
	return os.WriteFile(s.mortgageRequestsFile, data, 0644)
} 