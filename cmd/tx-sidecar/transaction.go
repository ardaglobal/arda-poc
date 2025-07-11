package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	zlog "github.com/rs/zerolog/log"
)

// TransactionEvent represents a single event in the transaction lifecycle
type TransactionEvent struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Height    int64                  `json:"height,omitempty"`
	Code      uint32                 `json:"code,omitempty"`
	RawLog    string                 `json:"raw_log,omitempty"`
	Events    []map[string]interface{} `json:"events,omitempty"`
}

// TrackedTx stores information about a transaction that has been broadcast.
type TrackedTx struct {
	Timestamp time.Time           `json:"timestamp"`
	Type      string              `json:"type"`
	TxHash    string              `json:"tx_hash"`
	Status    string              `json:"status"`
	Events    []TransactionEvent  `json:"events"`
}

// buildSignAndBroadcast handles the common logic for creating, signing, and broadcasting a transaction.
func (s *Server) buildSignAndBroadcast(w http.ResponseWriter, r *http.Request, fromName, txType string, msgBuilder func(fromAddr string) sdk.Msg) {
	txHash, err := s.buildSignAndBroadcastInternal(r.Context(), fromName, txType, msgBuilder)
	if err != nil {
		zlog.Error().Err(err).Msg("failed to build sign and broadcast")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"tx_hash": txHash})
}

// buildSignAndBroadcastInternal handles the core logic for creating, signing, and broadcasting a transaction
// without being tied to an HTTP handler.
func (s *Server) buildSignAndBroadcastInternal(ctx context.Context, fromName, txType string, msgBuilder func(fromAddr string) sdk.Msg) (string, error) {
	clientCtx := s.clientCtx
	// 1. Set the signer
	fromAddrRec, err := clientCtx.Keyring.Key(fromName)
	if err != nil {
		return "", fmt.Errorf("failed to get key for '%s': %w", fromName, err)
	}

	fromAddress, err := fromAddrRec.GetAddress()
	if err != nil {
		return "", fmt.Errorf("failed to get address from key: %w", err)
	}
	clientCtx = clientCtx.WithFrom(fromName).WithFromAddress(fromAddress)

	// 2. Create the message
	msg := msgBuilder(fromAddress.String())

	// 3. Build, sign, and broadcast
	txf := tx.Factory{}.
		WithChainID(clientCtx.ChainID).
		WithKeybase(clientCtx.Keyring).
		WithTxConfig(clientCtx.TxConfig).
		WithGas(200000) // Set a dummy gas limit

	acc, err := s.authClient.Account(ctx, &authtypes.QueryAccountRequest{Address: fromAddress.String()})
	if err != nil {
		return "", fmt.Errorf("failed to get account: %w", err)
	}

	var accI authtypes.AccountI
	if err := clientCtx.InterfaceRegistry.UnpackAny(acc.Account, &accI); err != nil {
		return "", fmt.Errorf("failed to unpack account into interface: %w", err)
	}

	baseAcc, ok := accI.(*authtypes.BaseAccount)
	if !ok {
		return "", fmt.Errorf("account is not a BaseAccount")
	}

	txf = txf.WithAccountNumber(baseAcc.AccountNumber).WithSequence(baseAcc.Sequence)

	txb, err := txf.BuildUnsignedTx(msg)
	if err != nil {
		return "", fmt.Errorf("failed to build unsigned tx: %w", err)
	}

	err = tx.Sign(ctx, txf, fromName, txb, true)
	if err != nil {
		return "", fmt.Errorf("failed to sign tx: %w", err)
	}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(txb.GetTx())
	if err != nil {
		return "", fmt.Errorf("failed to encode tx: %w", err)
	}

	// Use SYNC broadcast mode and then poll for the transaction to be included in a block.
	res, err := s.txClient.BroadcastTx(
		ctx,
		&txtypes.BroadcastTxRequest{
			Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes,
		},
	)
	if err != nil {
		s.trackTransactionEvent(txType, "", "failed", 0, err.Error(), nil)
		return "", fmt.Errorf("failed to broadcast tx: %w", err)
	}

	// In sync mode, a non-zero code means the transaction failed validation (CheckTx).
	if res.TxResponse.Code != 0 {
		s.trackTransactionEvent(txType, res.TxResponse.TxHash, "failed", res.TxResponse.Code, res.TxResponse.RawLog, nil)
		return "", fmt.Errorf("transaction failed with code %d: %s", res.TxResponse.Code, res.TxResponse.RawLog)
	}

	// Poll for the transaction to be included in a block.
	txHash := res.TxResponse.TxHash
	zlog.Info().Msgf("Transaction broadcasted with hash: %s. Polling for confirmation...", txHash)

	// Track the submitted state
	s.trackTransactionEvent(txType, txHash, "submitted", 0, "", nil)

	// This is a simplified polling mechanism. In a production system, you might want
	// a more robust solution, possibly involving a message queue or a dedicated transaction tracker.
	pollCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	for {
		txRes, err := s.txClient.GetTx(pollCtx, &txtypes.GetTxRequest{Hash: txHash})
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				// Transaction not yet in a block, wait and retry.
				time.Sleep(2 * time.Second)
				continue
			}
			s.trackTransactionEvent(txType, txHash, "failed", 0, fmt.Sprintf("failed to poll for tx confirmation: %v", err), nil)
			return "", fmt.Errorf("failed to poll for tx confirmation: %w", err)
		}

		// Transaction is confirmed.
		zlog.Info().Msgf("Transaction %s confirmed in block %d.", txHash, txRes.TxResponse.Height)
		
		// Extract events from the transaction response
		events := make([]map[string]interface{}, 0)
		for _, event := range txRes.TxResponse.Events {
			eventMap := make(map[string]interface{})
			eventMap["type"] = event.Type
			attributes := make([]map[string]string, 0)
			for _, attr := range event.Attributes {
				attributes = append(attributes, map[string]string{
					"key":   attr.Key,
					"value": attr.Value,
				})
			}
			eventMap["attributes"] = attributes
			events = append(events, eventMap)
		}

		if txRes.TxResponse.Code == 0 {
			s.trackTransactionEvent(txType, txHash, "confirmed", 0, "", events)
		} else {
			s.trackTransactionEvent(txType, txHash, "failed", txRes.TxResponse.Code, txRes.TxResponse.RawLog, events)
		}
		
		return txHash, nil
	}
}

// trackTransactionEvent adds a new transaction event to the server's list and saves it to a file.
func (s *Server) trackTransactionEvent(txType, txHash, status string, code uint32, rawLog string, events []map[string]interface{}) {
	now := time.Now()
	
	// Find existing transaction or create new one
	var existingTx *TrackedTx
	for i := range s.transactions {
		if s.transactions[i].TxHash == txHash {
			existingTx = &s.transactions[i]
			break
		}
	}
	
	if existingTx == nil {
		// Create new transaction
		newTx := TrackedTx{
			Timestamp: now,
			Type:      txType,
			TxHash:    txHash,
			Status:    status,
			Events:    make([]TransactionEvent, 0),
		}
		s.transactions = append(s.transactions, newTx)
		existingTx = &s.transactions[len(s.transactions)-1]
	}
	
	// Update transaction status and add event
	existingTx.Status = status
	event := TransactionEvent{
		Status:    status,
		Timestamp: now,
		Code:      code,
		RawLog:    rawLog,
		Events:    events,
	}
	existingTx.Events = append(existingTx.Events, event)
	
	s.saveTransactionsToFile()
}

// trackTransaction adds a new transaction to the server's list and saves it to a file.
// This method is kept for backward compatibility
func (s *Server) trackTransaction(txType, txHash string) {
	s.trackTransactionEvent(txType, txHash, "legacy", 0, "", nil)
}

// saveTransactionsToFile saves the current list of transactions to tx.json.
func (s *Server) saveTransactionsToFile() {
	data, err := json.MarshalIndent(s.transactions, "", "  ")
	if err != nil {
		zlog.Warn().Msgf("failed to marshal transactions: %v", err)
		return
	}
	if err := os.WriteFile(s.transactionsFile, data, 0644); err != nil {
		zlog.Warn().Msgf("failed to write transactions file: %v", err)
	}
}

// PaginatedTransactionsResponse represents the paginated response for transactions
type PaginatedTransactionsResponse struct {
	Transactions []TrackedTx `json:"transactions"`
	Total        int         `json:"total"`
	Page         int         `json:"page"`
	PageSize     int         `json:"page_size"`
	HasNext      bool        `json:"has_next"`
	HasPrev      bool        `json:"has_prev"`
}

// listTransactionsHandler returns the list of tracked transactions with pagination
// @Summary List transactions
// @Description Lists all transaction hashes that have been successfully processed and stored by the sidecar with pagination support.
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Number of transactions per page (default: 50, max: 1000)"
// @Success 200 {object} PaginatedTransactionsResponse
// @Router /tx/list [get]
func (s *Server) listTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	zlog.Info().Str("handler", "listTransactionsHandler").Msg("received request")
	
	// Parse pagination parameters
	page := 1
	pageSize := 50
	
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
			if pageSize > 1000 {
				pageSize = 1000 // Cap at 1000
			}
		}
	}
	
	total := len(s.transactions)
	totalPages := (total + pageSize - 1) / pageSize
	
	// Calculate start and end indices
	start := (page - 1) * pageSize
	end := start + pageSize
	
	var paginatedTransactions []TrackedTx
	if start < total {
		if end > total {
			end = total
		}
		paginatedTransactions = s.transactions[start:end]
	} else {
		paginatedTransactions = []TrackedTx{}
	}
	
	response := PaginatedTransactionsResponse{
		Transactions: paginatedTransactions,
		Total:        total,
		Page:         page,
		PageSize:     pageSize,
		HasNext:      page < totalPages,
		HasPrev:      page > 1,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getTransactionHandler returns a specific transaction by its hash
// @Summary Get transaction
// @Description Queries the blockchain for a specific transaction by its hash and returns details. For certain transaction types like 'register_property', it returns a richer, decoded response.
// @Produce json
// @Param hash path string true "Transaction hash"
// @Success 200 {object} interface{}
// @Router /tx/{hash} [get]
func (s *Server) getTransactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	txHash := strings.TrimPrefix(r.URL.Path, "/tx/")
	zlog.Info().Str("handler", "getTransactionHandler").Str("tx_hash", txHash).Msg("received request")
	if txHash == "" {
		http.Error(w, "Transaction hash must be provided in the path", http.StatusBadRequest)
		return
	}

	// Find our internal transaction type from the cache.
	var trackedTx TrackedTx
	found := false
	for _, tx := range s.transactions {
		if tx.TxHash == txHash {
			trackedTx = tx
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Transaction not found in local cache", http.StatusNotFound)
		return
	}

	// Query the blockchain for the full transaction details
	getTxRes, err := s.txClient.GetTx(r.Context(), &txtypes.GetTxRequest{Hash: txHash})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get transaction from blockchain: %v", err), http.StatusNotFound)
		return
	}

	if getTxRes.TxResponse.Code != 0 {
		http.Error(w, fmt.Sprintf("Transaction failed on-chain with code %d: %s", getTxRes.TxResponse.Code, getTxRes.TxResponse.RawLog), http.StatusInternalServerError)
		return
	}

	// Process the response based on the transaction type from tx.json
	switch trackedTx.Type {
	case "register_property", "transfer_shares", "edit_property_metadata", "create_mortgage", "repay_mortgage", "request_funds":
		// Build a richer response object, modeled after the provided txout.json example.
		response := make(map[string]interface{})

		// Add core TxResponse fields, ensuring height is a string.
		response["height"] = strconv.FormatInt(getTxRes.TxResponse.Height, 10)
		response["txhash"] = getTxRes.TxResponse.TxHash
		response["timestamp"] = getTxRes.TxResponse.Timestamp

		// Filter for and include only the 'submission' event, and stringify it.
		var submissionEvents sdk.StringEvents
		for _, event := range getTxRes.TxResponse.Events {
			if event.Type == "submission" {
				submissionEvents = append(submissionEvents, sdk.StringifyEvent(event))
			}
		}
		response["events"] = submissionEvents

		// Decode the transaction from the response to access its messages.
		sdkTx, err := s.clientCtx.TxConfig.TxDecoder()(getTxRes.TxResponse.Tx.Value)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to decode tx: %v", err), http.StatusInternalServerError)
			return
		}

		// Marshal each message into its JSON representation.
		var messages []json.RawMessage
		for _, msg := range sdkTx.GetMsgs() {
			jsonBytes, err := s.clientCtx.Codec.MarshalJSON(msg)
			if err != nil {
				zlog.Warn().Msgf("failed to marshal message to JSON: %v", err)
				http.Error(w, "Failed to marshal a transaction message to JSON", http.StatusInternalServerError)
				return
			}
			messages = append(messages, json.RawMessage(jsonBytes))
		}

		// Construct the "tx" object with the message bodies.
		response["tx"] = map[string]interface{}{
			"body": map[string]interface{}{
				"messages": messages,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}

	default:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(getTxRes.TxResponse)
	}
}

// getTransactionEventsHandler returns the lifecycle events for a specific transaction
// @Summary Get transaction events
// @Description Returns the complete lifecycle events for a transaction including submitted, confirmed, and failed states. This endpoint provides real-time updates on transaction status for frontend UI updates.
// @Produce json
// @Param hash path string true "Transaction hash"
// @Success 200 {object} TrackedTx
// @Failure 404 {object} map[string]string
// @Router /tx/events/{hash} [get]
func (s *Server) getTransactionEventsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	txHash := strings.TrimPrefix(r.URL.Path, "/tx/events/")
	zlog.Info().Str("handler", "getTransactionEventsHandler").Str("tx_hash", txHash).Msg("received request")
	
	if txHash == "" {
		http.Error(w, "Transaction hash must be provided in the path", http.StatusBadRequest)
		return
	}

	// Find the transaction in our local cache
	var trackedTx *TrackedTx
	for i := range s.transactions {
		if s.transactions[i].TxHash == txHash {
			trackedTx = &s.transactions[i]
			break
		}
	}

	if trackedTx == nil {
		http.Error(w, "Transaction not found in local cache", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trackedTx)
}
