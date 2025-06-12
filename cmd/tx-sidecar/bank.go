package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"cosmossdk.io/math"
	mortgagetypes "github.com/ardaglobal/arda-poc/x/mortgage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

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

	s.buildSignAndBroadcast(w, r, fromName, req.Gas, "create_mortgage", msgBuilder)
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