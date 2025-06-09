package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	sidecarclient "github.com/ardaglobal/arda-poc/pkg/client"
	propertytypes "github.com/ardaglobal/arda-poc/x/property/types"
)

// Server holds the dependencies for the sidecar http server.
type Server struct {
	clientCtx  client.Context
	authClient authtypes.QueryClient
	txClient   txtypes.ServiceClient
	users      map[string]UserData
	usersFile  string
}

// UserData holds the information for a created user.
type UserData struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Mnemonic string `json:"mnemonic"`
}

// NewServer creates a new instance of the Server with all its dependencies.
func NewServer(clientCtx client.Context, grpcAddr string) (*Server, error) {
	grpcConn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server at %s: %w", grpcAddr, err)
	}

	usersFile := "users.json"
	users := make(map[string]UserData)

	file, err := os.ReadFile(usersFile)
	if err == nil {
		if err := json.Unmarshal(file, &users); err != nil {
			log.Printf("Warning: failed to unmarshal users file, starting with empty user map: %v", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read users file: %w", err)
	}

	return &Server{
		clientCtx:  clientCtx,
		authClient: authtypes.NewQueryClient(grpcConn),
		txClient:   txtypes.NewServiceClient(grpcConn),
		users:      users,
		usersFile:  usersFile,
	}, nil
}

// Close is a no-op for this server version but can be used for cleanup.
func (s *Server) Close() {}

// Request body structure
type RegisterPropertyRequest struct {
	Address string   `json:"address"`
	Region  string   `json:"region"`
	Value   uint64   `json:"value"`
	Owners  []string `json:"owners"`
	Shares  []uint64 `json:"shares"`
}

// CreateUserRequest defines the request body for creating a new user.
type CreateUserRequest struct {
	Name string `json:"name"`
}

// KeyInfo defines the structure for returning key information.
type KeyInfo struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Address string `json:"address"`
	PubKey  string `json:"pubkey"`
}

// TransferSharesRequest defines the request body for transferring property shares.
type TransferSharesRequest struct {
	PropertyID string   `json:"property_id"`
	FromOwners []string `json:"from_owners"`
	FromShares []uint64 `json:"from_shares"`
	ToOwners   []string `json:"to_owners"`
	ToShares   []uint64 `json:"to_shares"`
}

// corsHandler wraps a handler to include CORS headers.
func corsHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle pre-flight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func main() {
	// This context is for the main application, not for individual requests.
	clientCtx, err := sidecarclient.NewClientContext()
	if err != nil {
		log.Fatalf("Failed to create client context: %v", err)
	}

	parsedURL, err := url.Parse(clientCtx.NodeURI)
	if err != nil {
		log.Fatalf("Failed to parse node URI: %v", err)
	}
	host := strings.Split(parsedURL.Host, ":")[0]
	grpcAddr := fmt.Sprintf("%s:9090", host)

	server, err := NewServer(clientCtx, grpcAddr)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer server.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/register-property", server.registerPropertyHandler)
	mux.HandleFunc("/transfer-shares", server.transferSharesHandler)
	mux.HandleFunc("/keys", server.listKeysHandler)
	mux.HandleFunc("/create-user", server.createUserHandler)

	fmt.Println("Starting transaction sidecar server on :8080...")
	if err := http.ListenAndServe(":8080", corsHandler(mux)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func (s *Server) registerPropertyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterPropertyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	clientCtx := s.clientCtx
	// 2. Set the signer
	fromName := "ERES" // In a real app, this might come from the request or config
	fromAddr, err := clientCtx.Keyring.Key(fromName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get key for '%s'", fromName), http.StatusInternalServerError)
		return
	}

	fromAddress, err := fromAddr.GetAddress()
	if err != nil {
		http.Error(w, "Failed to get address from key", http.StatusInternalServerError)
		return
	}
	clientCtx = clientCtx.WithFrom(fromName).WithFromAddress(fromAddress)

	// 3. Create the message
	msg := propertytypes.NewMsgRegisterProperty(
		clientCtx.GetFromAddress().String(),
		req.Address,
		req.Region,
		req.Value,
		req.Owners,
		req.Shares,
	)
	if err := msg.ValidateBasic(); err != nil {
		http.Error(w, fmt.Sprintf("Message validation failed: %v", err), http.StatusBadRequest)
		return
	}

	// 4. Build, sign, and broadcast
	txf := tx.Factory{}.
		WithChainID(clientCtx.ChainID).
		WithKeybase(clientCtx.Keyring).
		WithGas(200000).
		WithTxConfig(clientCtx.TxConfig)

	acc, err := s.authClient.Account(context.Background(), &authtypes.QueryAccountRequest{Address: fromAddress.String()})
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

	res, err := s.txClient.BroadcastTx(
		context.Background(),
		&txtypes.BroadcastTxRequest{
			Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes,
		},
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to broadcast tx: %v", err), http.StatusInternalServerError)
		return
	}

	// 5. Return the transaction hash
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"tx_hash": res.TxResponse.TxHash,
	})
	fmt.Printf("Successfully broadcasted tx with hash: %s\n", res.TxResponse.TxHash)
}

func (s *Server) createUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "Name cannot be empty", http.StatusBadRequest)
		return
	}

	// Check if key with this name already exists in the keyring
	if _, err := s.clientCtx.Keyring.Key(req.Name); err == nil {
		http.Error(w, fmt.Sprintf("User with name '%s' already exists", req.Name), http.StatusConflict)
		return
	}

	// Create a new key in the keyring
	record, mnemonic, err := s.clientCtx.Keyring.NewMnemonic(
		req.Name,
		keyring.English,
		sdk.GetConfig().GetFullBIP44Path(),
		keyring.DefaultBIP39Passphrase,
		hd.Secp256k1,
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create new key: %v", err), http.StatusInternalServerError)
		return
	}

	addr, err := record.GetAddress()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get address from record: %v", err), http.StatusInternalServerError)
		return
	}

	// Store user data in memory and save to file
	userData := UserData{
		Name:     req.Name,
		Address:  addr.String(),
		Mnemonic: mnemonic,
	}
	s.users[req.Name] = userData
	if err := s.saveUsersToFile(); err != nil {
		log.Printf("Warning: failed to save users to file: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(userData); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func (s *Server) saveUsersToFile() error {
	data, err := json.MarshalIndent(s.users, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal users: %w", err)
	}
	return os.WriteFile(s.usersFile, data, 0644)
}

func (s *Server) listKeysHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	records, err := s.clientCtx.Keyring.List()
	if err != nil {
		http.Error(w, "Failed to list keys from keyring", http.StatusInternalServerError)
		return
	}

	keyInfos := make([]KeyInfo, len(records))
	for i, record := range records {
		addr, err := record.GetAddress()
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get address for key '%s': %v", record.Name, err), http.StatusInternalServerError)
			return
		}

		pubKeyJSON, err := s.clientCtx.Codec.MarshalJSON(record.PubKey)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to marshal pubkey for key '%s': %v", record.Name, err), http.StatusInternalServerError)
			return
		}

		keyInfos[i] = KeyInfo{
			Name:    record.Name,
			Type:    "local",
			Address: addr.String(),
			PubKey:  string(pubKeyJSON),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(keyInfos); err != nil {
		http.Error(w, "Failed to encode keys to JSON", http.StatusInternalServerError)
		return
	}
}

func (s *Server) transferSharesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req TransferSharesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	clientCtx := s.clientCtx
	// In a real app, this might come from the request or config
	fromName := "ERES"
	fromAddr, err := clientCtx.Keyring.Key(fromName)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get key for '%s'", fromName), http.StatusInternalServerError)
		return
	}

	fromAddress, err := fromAddr.GetAddress()
	if err != nil {
		http.Error(w, "Failed to get address from key", http.StatusInternalServerError)
		return
	}
	clientCtx = clientCtx.WithFrom(fromName).WithFromAddress(fromAddress)

	// Create the message
	msg := propertytypes.NewMsgTransferShares(
		clientCtx.GetFromAddress().String(),
		req.PropertyID,
		req.FromOwners,
		req.FromShares,
		req.ToOwners,
		req.ToShares,
	)
	if err := msg.ValidateBasic(); err != nil {
		http.Error(w, fmt.Sprintf("Message validation failed: %v", err), http.StatusBadRequest)
		return
	}

	// Build, sign, and broadcast
	txf := tx.Factory{}.
		WithChainID(clientCtx.ChainID).
		WithKeybase(clientCtx.Keyring).
		WithGas(200000).
		WithTxConfig(clientCtx.TxConfig)

	acc, err := s.authClient.Account(context.Background(), &authtypes.QueryAccountRequest{Address: fromAddress.String()})
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

	res, err := s.txClient.BroadcastTx(
		context.Background(),
		&txtypes.BroadcastTxRequest{
			Mode:    txtypes.BroadcastMode_BROADCAST_MODE_SYNC,
			TxBytes: txBytes,
		},
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to broadcast tx: %v", err), http.StatusInternalServerError)
		return
	}

	// Return the transaction hash
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"tx_hash": res.TxResponse.TxHash,
	})
	fmt.Printf("Successfully broadcasted tx with hash: %s\n", res.TxResponse.TxHash)
}
