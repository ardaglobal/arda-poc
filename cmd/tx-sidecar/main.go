package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/math"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/yaml.v3"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	sidecarclient "github.com/ardaglobal/arda-poc/pkg/client"
	propertytypes "github.com/ardaglobal/arda-poc/x/property/types"
)

// Config structs for parsing config.yml
type FaucetConfig struct {
	Name string `yaml:"name"`
}

type AppConfig struct {
	Faucet FaucetConfig `yaml:"faucet"`
}

// Server holds the dependencies for the sidecar http server.
type Server struct {
	clientCtx    client.Context
	authClient   authtypes.QueryClient
	txClient     txtypes.ServiceClient
	users        map[string]UserData
	usersFile    string
	logins       map[string]string // email -> name
	loginsFile   string
	faucetName   string
	loggedInUser string
}

// UserData holds the information for a created user.
type UserData struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Mnemonic string `json:"mnemonic"`
	Role     string `json:"role"`
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

	// Sync keyring with users.json
	records, err := clientCtx.Keyring.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list keys from keyring: %w", err)
	}

	usersFileNeedsSave := false
	for _, record := range records {
		if _, ok := users[record.Name]; !ok {
			// User exists in keyring but not in users.json, add them.
			addr, err := record.GetAddress()
			if err != nil {
				log.Printf("Warning: failed to get address for key '%s', skipping sync: %v", record.Name, err)
				continue
			}
			users[record.Name] = UserData{
				Name:     record.Name,
				Address:  addr.String(),
				Mnemonic: "", // Mnemonic is not available from keyring listing
				Role:     "user",
			}
			log.Printf("Syncing key '%s' from keyring to users.json", record.Name)
			usersFileNeedsSave = true
		}
	}

	if usersFileNeedsSave {
		data, err := json.MarshalIndent(users, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to marshal users for sync: %w", err)
		}
		if err := os.WriteFile(usersFile, data, 0644); err != nil {
			return nil, fmt.Errorf("failed to write users file for sync: %w", err)
		}
	}

	loginsFile := "logins.json"
	logins := make(map[string]string)

	loginData, err := os.ReadFile(loginsFile)
	if err == nil {
		if err := json.Unmarshal(loginData, &logins); err != nil {
			log.Printf("Warning: failed to unmarshal logins file, starting with empty login map: %v", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read logins file: %w", err)
	}

	// Read faucet configuration
	configPath := "config.yml"
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var appConfig AppConfig
	if err := yaml.Unmarshal(configData, &appConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file: %w", err)
	}

	if appConfig.Faucet.Name == "" {
		return nil, fmt.Errorf("faucet name is not defined in config.yml")
	}

	s := &Server{
		clientCtx:  clientCtx,
		authClient: authtypes.NewQueryClient(grpcConn),
		txClient:   txtypes.NewServiceClient(grpcConn),
		users:      users,
		usersFile:  usersFile,
		logins:     logins,
		loginsFile: loginsFile,
		faucetName: appConfig.Faucet.Name,
	}

	// Ensure that the faucet account from config exists in the keyring.
	if _, err := s.clientCtx.Keyring.Key(s.faucetName); err != nil {
		return nil, fmt.Errorf("faucet user '%s' from config.yml not found in keyring: %w", s.faucetName, err)
	}
	log.Printf("Using '%s' as the faucet account.", s.faucetName)

	// Ensure faucet user has the 'faucet' role.
	if faucetUserData, ok := s.users[s.faucetName]; ok {
		if faucetUserData.Role != "faucet" {
			log.Printf("Updating role of faucet user '%s' to 'faucet'.", s.faucetName)
			faucetUserData.Role = "faucet"
			s.users[s.faucetName] = faucetUserData
			if err := s.saveUsersToFile(); err != nil {
				log.Printf("Warning: failed to save users file after updating faucet role: %v", err)
			}
		}
	}

	return s, nil
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
	Gas     string   `json:"gas,omitempty"`
}

// LoginRequest defines the request body for logging in or registering a user.
type LoginRequest struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
	Role  string `json:"role,omitempty"`
}

// LoginResponse defines the structure of the response for the login endpoint.
type LoginResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	User    string `json:"user"`
	Role    string `json:"role,omitempty"`
}

// UserDetailResponse defines the structure for returning detailed user information.
type UserDetailResponse struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Role    string `json:"role"`
	Type    string `json:"type"`
	PubKey  string `json:"pubkey"`
}

// TransferSharesRequest defines the request body for transferring property shares.
type TransferSharesRequest struct {
	PropertyID string   `json:"property_id"`
	FromOwners []string `json:"from_owners"`
	FromShares []uint64 `json:"from_shares"`
	ToOwners   []string `json:"to_owners"`
	ToShares   []uint64 `json:"to_shares"`
	Gas        string   `json:"gas,omitempty"`
}

// FaucetRequest defines the request body for requesting funds from the faucet.
type FaucetRequest struct {
	Address string `json:"address"`
	Amount  uint64 `json:"amount"`
	Denom   string `json:"denom"`
	Gas     string `json:"gas,omitempty"`
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
	mux.HandleFunc("/users", server.listUsersHandler)
	mux.HandleFunc("/login", server.loginHandler)
	mux.HandleFunc("/logout", server.logoutHandler)
	mux.HandleFunc("/faucet", server.faucetHandler)

	fmt.Println("Starting transaction sidecar server on :8080...")
	if err := http.ListenAndServe(":8080", corsHandler(mux)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// buildSignAndBroadcast handles the common logic for creating, signing, and broadcasting a transaction.
func (s *Server) buildSignAndBroadcast(w http.ResponseWriter, r *http.Request, fromName string, gasStr string, msgBuilder func(fromAddr string) sdk.Msg) {
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
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]string{
						"tx_hash": getTxRes.TxResponse.TxHash,
					})
					fmt.Printf("Successfully processed tx with hash: %s\n", getTxRes.TxResponse.TxHash)
				}
				return // Exit polling.
			}

			// Continue polling if the transaction is not yet found.
			time.Sleep(1 * time.Second)
		}
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

	fromName := "ERES" // In a real app, this might come from the request or config
	msgBuilder := func(fromAddr string) sdk.Msg {
		return propertytypes.NewMsgRegisterProperty(
			fromAddr,
			req.Address,
			req.Region,
			req.Value,
			req.Owners,
			req.Shares,
		)
	}

	s.buildSignAndBroadcast(w, r, fromName, req.Gas, msgBuilder)
}

func (s *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" {
		http.Error(w, "Email cannot be empty", http.StatusBadRequest)
		return
	}

	name, emailExists := s.logins[req.Email]

	if s.loggedInUser != "" {
		if emailExists && name == s.loggedInUser {
			w.Header().Set("Content-Type", "application/json")
			userData := s.users[s.loggedInUser]
			json.NewEncoder(w).Encode(LoginResponse{
				Status:  "success",
				Message: fmt.Sprintf("User %s is already logged in", s.loggedInUser),
				User:    s.loggedInUser,
				Role:    userData.Role,
			})
			return
		}
		http.Error(w, fmt.Sprintf("User '%s' is already logged in. Please log out first.", s.loggedInUser), http.StatusConflict)
		return
	}

	// from here, we know s.loggedInUser == ""

	if emailExists {
		s.loggedInUser = name
		userData, ok := s.users[name]
		if !ok {
			http.Error(w, fmt.Sprintf("internal data inconsistency: user %s not found", name), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(LoginResponse{
			Status:  "success",
			Message: fmt.Sprintf("User %s logged in", name),
			User:    name,
			Role:    userData.Role,
		})
		return
	}

	// Email doesn't exist. This is a registration/linking flow.
	if req.Name == "" {
		http.Error(w, "Email not registered. Please provide a name to create a new user.", http.StatusBadRequest)
		return
	}

	var finalUserData UserData
	// Check if the user `name` already exists in the keyring
	_, err := s.clientCtx.Keyring.Key(req.Name)
	if err != nil { // User does not exist, create new one
		createdUser, err := s.createUser(req.Name, req.Role)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create user: %v", err), http.StatusBadRequest)
			return
		}
		finalUserData = *createdUser
		log.Printf("Created new user '%s' with address %s and role %s", finalUserData.Name, finalUserData.Address, finalUserData.Role)
	} else {
		existingUser, ok := s.users[req.Name]
		if !ok {
			http.Error(w, fmt.Sprintf("internal data inconsistency: user %s not found", req.Name), http.StatusInternalServerError)
			return
		}
		finalUserData = existingUser
		log.Printf("User with name '%s' already exists, linking to email '%s'", req.Name, req.Email)
	}

	// Map email to name and save
	s.logins[req.Email] = req.Name
	if err := s.saveLoginsToFile(); err != nil {
		log.Printf("Warning: failed to save logins to file: %v", err)
	}

	s.loggedInUser = req.Name
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(LoginResponse{
		Status:  "success",
		Message: fmt.Sprintf("User %s created/linked and logged in", req.Name),
		User:    req.Name,
		Role:    finalUserData.Role,
	})
}

func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if s.loggedInUser == "" {
		http.Error(w, "No user is currently logged in", http.StatusBadRequest)
		return
	}

	loggedOutUser := s.loggedInUser
	s.loggedInUser = ""
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("User %s logged out", loggedOutUser),
	})
}

func (s *Server) createUser(name, role string) (*UserData, error) {
	// Check if key with this name already exists in the keyring
	if _, err := s.clientCtx.Keyring.Key(name); err == nil {
		return nil, fmt.Errorf("user with name '%s' already exists", name)
	}

	// Create a new key in the keyring
	record, mnemonic, err := s.clientCtx.Keyring.NewMnemonic(
		name,
		keyring.English,
		sdk.GetConfig().GetFullBIP44Path(),
		keyring.DefaultBIP39Passphrase,
		hd.Secp256k1,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create new key: %v", err)
	}

	addr, err := record.GetAddress()
	if err != nil {
		return nil, fmt.Errorf("failed to get address from record: %v", err)
	}

	// Validate and set role
	finalRole := "user" // default role
	if role != "" {
		allowedRoles := map[string]bool{
			"user":      true,
			"investor":  true,
			"developer": true,
			"regulator": true,
			"admin":     true,
			"faucet":    true,
		}
		if _, ok := allowedRoles[role]; !ok {
			return nil, fmt.Errorf("invalid role provided: '%s'. aRole must be one of user, investor, developer, regulator, admin, or faucet", role)
		}
		finalRole = role
	}

	// Store user data in memory and save to file
	userData := UserData{
		Name:     name,
		Address:  addr.String(),
		Mnemonic: mnemonic,
		Role:     finalRole,
	}
	s.users[name] = userData
	if err := s.saveUsersToFile(); err != nil {
		log.Printf("Warning: failed to save users to file: %v", err)
	}

	return &userData, nil
}

func (s *Server) saveUsersToFile() error {
	data, err := json.MarshalIndent(s.users, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal users: %w", err)
	}
	return os.WriteFile(s.usersFile, data, 0644)
}

func (s *Server) saveLoginsToFile() error {
	data, err := json.MarshalIndent(s.logins, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal logins: %w", err)
	}
	return os.WriteFile(s.loginsFile, data, 0644)
}

func (s *Server) listUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	userInfos := make([]UserDetailResponse, 0, len(s.users))
	for name, userData := range s.users {
		record, err := s.clientCtx.Keyring.Key(name)
		if err != nil {
			log.Printf("Warning: User '%s' is in users.json but not in the keyring. Listing without key info.", name)
			userInfos = append(userInfos, UserDetailResponse{
				Name:    userData.Name,
				Address: userData.Address,
				Role:    userData.Role,
				Type:    "local (key missing)",
				PubKey:  "",
			})
			continue
		}

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

		userInfos = append(userInfos, UserDetailResponse{
			Name:    record.Name,
			Type:    "local",
			Address: addr.String(),
			PubKey:  string(pubKeyJSON),
			Role:    userData.Role,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(userInfos); err != nil {
		http.Error(w, "Failed to encode users to JSON", http.StatusInternalServerError)
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

	fromName := "ERES" // In a real app, this might come from the request or config
	msgBuilder := func(fromAddr string) sdk.Msg {
		return propertytypes.NewMsgTransferShares(
			fromAddr,
			req.PropertyID,
			req.FromOwners,
			req.FromShares,
			req.ToOwners,
			req.ToShares,
		)
	}

	s.buildSignAndBroadcast(w, r, fromName, req.Gas, msgBuilder)
}

func (s *Server) faucetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req FaucetRequest
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

	s.buildSignAndBroadcast(w, r, fromName, req.Gas, msgBuilder)
}
