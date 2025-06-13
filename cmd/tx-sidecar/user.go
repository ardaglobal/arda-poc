package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	zlog "github.com/rs/zerolog/log"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
)

// UserData holds the information for a created user.
type UserData struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Mnemonic string `json:"mnemonic"`
	Role     string `json:"role"`
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

// LoginStatusResponse defines the structure of the response for the login status endpoint.
type LoginStatusResponse struct {
	LoggedIn bool   `json:"logged_in"`
	User     string `json:"user,omitempty"`
	Role     string `json:"role,omitempty"`
}

// KYCRequest defines the request body for KYC'ing a user.
type KYCRequest struct {
	Name string `json:"name"`
}

// UserDetailResponse defines the structure for returning detailed user information.
type UserDetailResponse struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Role    string `json:"role"`
	Type    string `json:"type"`
	PubKey  string `json:"pubkey"`
}

// KYCRequestEntry defines a pending KYC request.
type KYCRequestEntry struct {
	ID        string    `json:"id"`
	Requester string    `json:"requester"` // Name of the user requesting KYC
	Status    string    `json:"status"`    // "pending", "approved", "rejected"
	Timestamp time.Time `json:"timestamp"`
}

// ApproveKYCRequest defines the request body for approving a KYC request.
type ApproveKYCRequest struct {
	ID string `json:"id"`
}

// KYCStatusResponse defines a standard status response for KYC endpoints.
type KYCStatusResponse struct {
	Status string `json:"status"`
}

// KYCErrorResponse defines a standard error response for KYC endpoints.
type KYCErrorResponse struct {
	Error string `json:"error"`
}

// loginHandler handles user login and registration
// @Summary User login, registration, and linking
// @Description Handles user login, registration, and linking. If a user with the given email exists, they are logged in. If the email does not exist and a name is provided, a new user account and key are created. If the email does not exist but a user with the given name does exist, the email is linked to the existing user account.
// @Accept json
// @Produce json
// @Param request body LoginRequest true "login info"
// @Success 200 {object} LoginResponse
// @Success 201 {object} LoginResponse
// @Router /user/login [post]
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

	zlog.Info().Str("handler", "loginHandler").Str("email", req.Email).Str("name", req.Name).Str("role", req.Role).Msg("received login request")

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
		zlog.Info().Msgf("Created new user '%s' with address %s and role %s", finalUserData.Name, finalUserData.Address, finalUserData.Role)
	} else {
		existingUser, ok := s.users[req.Name]
		if !ok {
			http.Error(w, fmt.Sprintf("internal data inconsistency: user %s not found", req.Name), http.StatusInternalServerError)
			return
		}
		finalUserData = existingUser
		zlog.Info().Msgf("User with name '%s' already exists, linking to email '%s'", req.Name, req.Email)
	}

	// Map email to name and save
	s.logins[req.Email] = req.Name
	if err := s.saveLoginsToFile(); err != nil {
		zlog.Warn().Msgf("failed to save logins to file: %v", err)
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

// logoutHandler logs out the current user
// @Summary User logout
// @Description Logs out the currently authenticated user.
// @Produce json
// @Success 200 {object} map[string]string{status=string,message=string}
// @Router /user/logout [post]
func (s *Server) logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	zlog.Info().Str("handler", "logoutHandler").Str("loggedInUser", s.loggedInUser).Msg("received logout request")

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

// loginStatusHandler returns the currently logged in user.
// @Summary Get login status
// @Description Returns the currently logged in user, if any.
// @Produce json
// @Success 200 {object} LoginStatusResponse
// @Router /user/status [get]
func (s *Server) loginStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	zlog.Info().Str("handler", "loginStatusHandler").Str("loggedInUser", s.loggedInUser).Msg("received request")

	w.Header().Set("Content-Type", "application/json")
	if s.loggedInUser == "" {
		json.NewEncoder(w).Encode(LoginStatusResponse{
			LoggedIn: false,
		})
		return
	}

	userData, ok := s.users[s.loggedInUser]
	if !ok {
		// This is an inconsistent state, log out the user
		oldUser := s.loggedInUser
		s.loggedInUser = ""
		zlog.Error().Msgf("data inconsistency: logged in user '%s' not found in user map. Forcing logout.", oldUser)
		http.Error(w, fmt.Sprintf("internal data inconsistency: logged in user '%s' not found", oldUser), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(LoginStatusResponse{
		LoggedIn: true,
		User:     s.loggedInUser,
		Role:     userData.Role,
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
			"bank":      true,
		}
		if _, ok := allowedRoles[role]; !ok {
			return nil, fmt.Errorf("invalid role provided: '%s'. aRole must be one of user, investor, developer, regulator, admin, or bank", role)
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
		zlog.Warn().Msgf("failed to save users to file: %v", err)
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

// listUsersHandler lists all users
// @Summary List users
// @Description Lists all registered users and their key details.
// @Produce json
// @Success 200 {array} UserDetailResponse
// @Router /user/list [get]
func (s *Server) listUsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	zlog.Info().Str("handler", "listUsersHandler").Msg("received request")

	userInfos := make([]UserDetailResponse, 0, len(s.users))
	for name, userData := range s.users {
		record, err := s.clientCtx.Keyring.Key(name)
		if err != nil {
			zlog.Warn().Msgf("User '%s' is in users.json but not in the keyring. Listing without key info.", name)
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

func (s *Server) initUsers() error {
	initialUsers := map[string]string{
		"matt":              "admin",
		"ERES":              "regulator",
		"DLD":               "regulator",
		"Fatima_Developers": "developer",
		"AIX":               "developer",
		"oli":               "investor",
		"toby":              "investor",
		"fawad":             "investor",
		"siddiq":            "investor",
	}

	for name, role := range initialUsers {
		if userData, ok := s.users[name]; !ok {
			zlog.Info().Msgf("User '%s' not found, creating...", name)
			_, err := s.createUser(name, role)
			if err != nil {
				return fmt.Errorf("failed to create initial user '%s': %w", name, err)
			}
			zlog.Info().Msgf("Successfully created initial user '%s'", name)
		} else {
			// User exists, check if the role is correct.
			if userData.Role == "faucet" {
				// Migrate old "faucet" role to "bank"
				zlog.Info().Msgf("Migrating user '%s' from deprecated 'faucet' role to 'bank'.", name)
				userData.Role = "bank"
				s.users[name] = userData
				if err := s.saveUsersToFile(); err != nil {
					return fmt.Errorf("failed to save users file after migrating %s's role: %w", name, err)
				}
			} else if userData.Role != role {
				zlog.Info().Msgf("User '%s' has incorrect role '%s', updating to '%s'.", name, userData.Role, role)
				userData.Role = role
				s.users[name] = userData
				if err := s.saveUsersToFile(); err != nil {
					return fmt.Errorf("failed to save users file after updating %s's role: %w", name, err)
				}
				zlog.Info().Msgf("Successfully updated role for user '%s'", name)
			}
		}
	}

	return nil
}

// requestKYCHandler allows a user to request KYC.
// @Summary Request KYC (user)
// @Description Allows a logged-in user to request KYC. This creates a pending KYC request that a regulator can later approve.
// @Accept json
// @Produce json
// @Success 201 {object} KYCRequestEntry
// @Failure 400 {object} KYCErrorResponse
// @Failure 401 {object} KYCErrorResponse
// @Router /user/kyc/request [post]
func (s *Server) requestKYCHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	if s.loggedInUser == "" {
		http.Error(w, "No user is logged in. Please log in to request KYC.", http.StatusUnauthorized)
		return
	}
	// Check if user already has a pending KYC request
	for _, req := range s.kycRequests {
		if req.Requester == s.loggedInUser && req.Status == "pending" {
			http.Error(w, "You already have a pending KYC request.", http.StatusBadRequest)
			return
		}
	}
	newReq := KYCRequestEntry{
		ID:        uuid.New().String(),
		Requester: s.loggedInUser,
		Status:    "pending",
		Timestamp: time.Now(),
	}
	s.kycRequests = append(s.kycRequests, newReq)
	if err := s.saveKYCRequestsToFile(); err != nil {
		http.Error(w, "Failed to save KYC request", http.StatusInternalServerError)
		zlog.Error().Err(err).Msg("failed to save KYC requests file")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newReq)
}

// getKYCRequestsHandler allows a regulator to view all pending KYC requests or a user to view their own pending request(s).
// @Summary Get pending KYC requests
// @Description Regulators see all pending KYC requests. Regular users see only their own pending KYC request(s).
// @Produce json
// @Success 200 {array} KYCRequestEntry
// @Failure 401 {object} KYCErrorResponse
// @Router /user/kyc/requests [get]
func (s *Server) getKYCRequestsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	if s.loggedInUser == "" {
		http.Error(w, "No user is logged in.", http.StatusUnauthorized)
		return
	}
	userData, ok := s.users[s.loggedInUser]
	if ok && userData.Role == "regulator" {
		// Regulator: see all pending requests
		pending := make([]KYCRequestEntry, 0)
		for _, req := range s.kycRequests {
			if req.Status == "pending" {
				pending = append(pending, req)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pending)
		return
	}
	// Non-regulator: see only their own pending request(s)
	userPending := make([]KYCRequestEntry, 0)
	for _, req := range s.kycRequests {
		if req.Status == "pending" && req.Requester == s.loggedInUser {
			userPending = append(userPending, req)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userPending)
}

// approveKYCHandler allows a regulator to approve a KYC request.
// @Summary Approve KYC request (regulator)
// @Description Allows a logged-in regulator to approve a pending KYC request. The user's role will be updated to 'investor'.
// @Accept json
// @Produce json
// @Param request body ApproveKYCRequest true "KYC approval request"
// @Success 200 {object} KYCStatusResponse
// @Failure 400 {object} KYCErrorResponse
// @Failure 401 {object} KYCErrorResponse
// @Failure 403 {object} KYCErrorResponse
// @Failure 404 {object} KYCErrorResponse
// @Router /user/kyc/approve [post]
func (s *Server) approveKYCHandler(w http.ResponseWriter, r *http.Request) {
	zlog.Info().Str("handler", "approveKYCHandler").Str("loggedInUser", s.loggedInUser).Fields(map[string]interface{}{
		"request": r.Body,
	}).Msg("received request")
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	if s.loggedInUser == "" {
		http.Error(w, "No user is logged in.", http.StatusUnauthorized)
		return
	}
	userData, ok := s.users[s.loggedInUser]
	if !ok || userData.Role != "regulator" {
		http.Error(w, "Only regulators can approve KYC requests.", http.StatusForbidden)
		return
	}
	var req ApproveKYCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	found := false
	for i, kycReq := range s.kycRequests {
		if kycReq.ID == req.ID && kycReq.Status == "pending" {
			s.kycRequests[i].Status = "approved"
			// Mark user as KYC'd (role = investor)
			requesterData, ok := s.users[kycReq.Requester]
			if ok && requesterData.Role == "user" {
				requesterData.Role = "investor"
				s.users[kycReq.Requester] = requesterData
				s.saveUsersToFile()
			}
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "KYC request not found or already processed", http.StatusNotFound)
		return
	}
	s.saveKYCRequestsToFile()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(KYCStatusResponse{Status: "approved"})
}

// Save/load KYC requests to/from file
func (s *Server) saveKYCRequestsToFile() error {
	data, err := json.MarshalIndent(s.kycRequests, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal KYC requests: %w", err)
	}
	return os.WriteFile(s.kycRequestsFile, data, 0644)
}

func (s *Server) loadKYCRequestsFromFile() error {
	data, err := os.ReadFile(s.kycRequestsFile)
	if err != nil {
		if os.IsNotExist(err) {
			s.kycRequests = []KYCRequestEntry{}
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &s.kycRequests)
}
