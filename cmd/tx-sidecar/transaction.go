package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// TrackedTx stores information about a transaction that has been broadcast.
type TrackedTx struct {
	Timestamp time.Time `json:"timestamp"`
	Type      string    `json:"type"`
	TxHash    string    `json:"tx_hash"`
}

// buildSignAndBroadcast handles the common logic for creating, signing, and broadcasting a transaction.
func (s *Server) buildSignAndBroadcast(w http.ResponseWriter, r *http.Request, fromName, gasStr, txType string, msgBuilder func(fromAddr string) sdk.Msg) {
	txHash, err := s.buildSignAndBroadcastInternal(r.Context(), fromName, gasStr, txType, msgBuilder)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"tx_hash": txHash})
}

// buildSignAndBroadcastInternal handles the core logic for creating, signing, and broadcasting a transaction
// without being tied to an HTTP handler.
func (s *Server) buildSignAndBroadcastInternal(ctx context.Context, fromName, gasStr, txType string, msgBuilder func(fromAddr string) sdk.Msg) (string, error) {
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
		WithGasAdjustment(2.0).
		WithGasPrices("0.025uarda") // NOTE: This may need to be configurable depending on the chain's requirements.

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

	var gas uint64 = 400000
	if gasStr != "" && gasStr != "auto" {
		parsedGas, err := strconv.ParseUint(gasStr, 10, 64)
		if err != nil {
			return "", fmt.Errorf("invalid gas value provided: %s", gasStr)
		}
		gas = parsedGas
	}
	txf = txf.WithGas(gas)

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
		return "", fmt.Errorf("failed to broadcast tx: %w", err)
	}

	// In sync mode, a non-zero code means the transaction failed validation (CheckTx).
	if res.TxResponse.Code != 0 {
		return "", fmt.Errorf("transaction failed with code %d: %s", res.TxResponse.Code, res.TxResponse.RawLog)
	}

	// Poll for the transaction to be included in a block.
	txHash := res.TxResponse.TxHash
	log.Printf("Transaction broadcasted with hash: %s. Polling for confirmation...", txHash)

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
			return "", fmt.Errorf("failed to poll for tx confirmation: %w", err)
		}

		// Transaction is confirmed.
		log.Printf("Transaction %s confirmed in block %d.", txHash, txRes.TxResponse.Height)
		s.trackTransaction(txType, txHash)
		return txHash, nil
	}
}

// trackTransaction adds a new transaction to the server's list and saves it to a file.
func (s *Server) trackTransaction(txType, txHash string) {
	newTx := TrackedTx{
		Timestamp: time.Now(),
		Type:      txType,
		TxHash:    txHash,
	}

	s.transactions = append(s.transactions, newTx)
	s.saveTransactionsToFile()
}

// saveTransactionsToFile saves the current list of transactions to tx.json.
func (s *Server) saveTransactionsToFile() {
	data, err := json.MarshalIndent(s.transactions, "", "  ")
	if err != nil {
		log.Printf("Warning: failed to marshal transactions: %v", err)
		return
	}
	if err := os.WriteFile(s.transactionsFile, data, 0644); err != nil {
		log.Printf("Warning: failed to write transactions file: %v", err)
	}
}

// listTransactionsHandler returns the list of tracked transactions.
func (s *Server) listTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.transactions)
}

// getTransactionHandler returns a specific transaction by its hash.
func (s *Server) getTransactionHandler(w http.ResponseWriter, r *http.Request) {
	txHash := strings.TrimPrefix(r.URL.Path, "/transaction/")
	if txHash == "" {
		http.Error(w, "Transaction hash must be provided", http.StatusBadRequest)
		return
	}

	for _, tx := range s.transactions {
		if tx.TxHash == txHash {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(tx)
			return
		}
	}

	http.Error(w, "Transaction not found", http.StatusNotFound)
} 