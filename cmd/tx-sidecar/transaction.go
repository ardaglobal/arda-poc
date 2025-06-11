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
	clientCtx := s.clientCtx
	// 1. Set the signer
	fromAddrRec, err := clientCtx.Keyring.Key(fromName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get key for '%s'", fromName), http.StatusInternalServerError)
		return
	}

	fromAddress, err := fromAddrRec.GetAddress()
	if err != nil {
		http.Error(w, "Failed to get address from key", http.StatusInternalServerError)
		return
	}
	clientCtx = clientCtx.WithFrom(fromName).WithFromAddress(fromAddress)

	// 2. Create the message
	msg := msgBuilder(fromAddress.String())

	// 3. Build, sign, and broadcast
	txf := tx.Factory{}.
		WithChainID(clientCtx.ChainID).
		WithKeybase(clientCtx.Keyring).
		WithTxConfig(clientCtx.TxConfig).
		WithGasAdjustment(1.5).
		WithGasPrices("0.025uarda") // NOTE: This may need to be configurable depending on the chain's requirements.

	acc, err := s.authClient.Account(r.Context(), &authtypes.QueryAccountRequest{Address: fromAddress.String()})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get account: %v", err), http.StatusInternalServerError)
		return
	}

	var accI authtypes.AccountI
	if err := clientCtx.InterfaceRegistry.UnpackAny(acc.Account, &accI); err != nil {
		http.Error(w, fmt.Sprintf("Failed to unpack account into interface: %v", err), http.StatusInternalServerError)
		return
	}

	baseAcc, ok := accI.(*authtypes.BaseAccount)
	if !ok {
		http.Error(w, "account is not a BaseAccount", http.StatusInternalServerError)
		return
	}

	txf = txf.WithAccountNumber(baseAcc.AccountNumber).WithSequence(baseAcc.Sequence)

	// We are removing auto gas calculation for now as it can be unreliable in this context.
	// Using a generous default. This can be overridden by the request if needed.
	var gas uint64 = 300000
	if gasStr != "" && gasStr != "auto" {
		parsedGas, err := strconv.ParseUint(gasStr, 10, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid gas value provided: %s", gasStr), http.StatusBadRequest)
			return
		}
		gas = parsedGas
	}
	txf = txf.WithGas(gas)

	txb, err := txf.BuildUnsignedTx(msg)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to build unsigned tx: %v", err), http.StatusInternalServerError)
		return
	}

	err = tx.Sign(r.Context(), txf, fromName, txb, true)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to sign tx: %v", err), http.StatusInternalServerError)
		return
	}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(txb.GetTx())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode tx: %v", err), http.StatusInternalServerError)
		return
	}

	// Use SYNC broadcast mode and then poll for the transaction to be included in a block.
	res, err := s.txClient.BroadcastTx(
		r.Context(),
		&txtypes.BroadcastTxRequest{
			Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes,
		},
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to broadcast tx: %v", err), http.StatusInternalServerError)
		return
	}

	// In sync mode, a non-zero code means the transaction failed validation (CheckTx).
	if res.TxResponse.Code != 0 {
		http.Error(w, fmt.Sprintf("Transaction failed on CheckTx with code %d: %s", res.TxResponse.Code, res.TxResponse.RawLog), http.StatusInternalServerError)
		return
	}

	// Poll for the transaction to be included in a block.
	txHash := res.TxResponse.TxHash
	pollCtx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	for {
		select {
		case <-pollCtx.Done():
			http.Error(w, fmt.Sprintf("Timed out waiting for transaction confirmation for hash %s. It may have failed or is still pending.", txHash), http.StatusInternalServerError)
			return
		default:
			getTxRes, err := s.txClient.GetTx(pollCtx, &txtypes.GetTxRequest{Hash: txHash})
			if err == nil {
				// Transaction found.
				if getTxRes.TxResponse.Code != 0 {
					// It was included in a block but failed during execution.
					http.Error(w, fmt.Sprintf("Transaction failed in block with code %d: %s", getTxRes.TxResponse.Code, getTxRes.TxResponse.RawLog), http.StatusInternalServerError)
				} else {
					// Success.
					txHash := getTxRes.TxResponse.TxHash
					s.addTransaction(txType, txHash)
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]string{
						"tx_hash": txHash,
					})
					fmt.Printf("Successfully processed tx with hash: %s\n", txHash)
				}
				return // Exit polling.
			}

			// Continue polling if the transaction is not yet found.
			time.Sleep(1 * time.Second)
		}
	}
}

func (s *Server) addTransaction(txType, txHash string) {
	tx := TrackedTx{
		Timestamp: time.Now().UTC(),
		Type:      txType,
		TxHash:    txHash,
	}
	s.transactions = append(s.transactions, tx)
	if err := s.saveTransactionsToFile(); err != nil {
		log.Printf("Warning: failed to save transactions to file: %v", err)
	}
}

func (s *Server) saveTransactionsToFile() error {
	data, err := json.MarshalIndent(s.transactions, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal transactions: %w", err)
	}
	return os.WriteFile(s.transactionsFile, data, 0644)
}

func (s *Server) listTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(s.transactions); err != nil {
		http.Error(w, "Failed to encode transactions to JSON", http.StatusInternalServerError)
		return
	}
}

func (s *Server) getTransactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	txHash := strings.TrimPrefix(r.URL.Path, "/transaction/")
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
	case "register_property", "transfer_shares":
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
				log.Printf("Warning: failed to marshal message to JSON: %v", err)
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
