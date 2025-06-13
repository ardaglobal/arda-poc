package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	fiber "github.com/gofiber/fiber/v2"

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
func (s *Server) loginHandler(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodPost {
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Invalid request method")
	}

	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
	}

	zlog.Info().Str("handler", "loginHandler").Str("email", req.Email).Str("name", req.Name).Str("role", req.Role).Msg("received login request")

	if req.Email == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Email cannot be empty")
	}

	name, emailExists := s.logins[req.Email]

	if s.loggedInUser != "" {
		if emailExists && name == s.loggedInUser {
			userData := s.users[s.loggedInUser]
			return c.JSON(LoginResponse{
				Status:  "success",
				Message: fmt.Sprintf("User %s is already logged in", s.loggedInUser),
				User:    s.loggedInUser,
				Role:    userData.Role,
			})
		}
		return c.Status(fiber.StatusConflict).SendString(fmt.Sprintf("User '%s' is already logged in. Please log out first.", s.loggedInUser))
	}

	// from here, we know s.loggedInUser == ""

	if emailExists {
		s.loggedInUser = name
		userData, ok := s.users[name]
		if !ok {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("internal data inconsistency: user %s not found", name))
		}
		return c.JSON(LoginResponse{
			Status:  "success",
			Message: fmt.Sprintf("User %s logged in", name),
			User:    name,
			Role:    userData.Role,
		})
	}

	// Email doesn't exist. This is a registration/linking flow.
	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Email not registered. Please provide a name to create a new user.")
	}

	var finalUserData UserData
	// Check if the user `name` already exists in the keyring
	_, err := s.clientCtx.Keyring.Key(req.Name)
	if err != nil { // User does not exist, create new one
		createdUser, err := s.createUser(req.Name, req.Role)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("Failed to create user: %v", err))
		}
		finalUserData = *createdUser
		zlog.Info().Msgf("Created new user '%s' with address %s and role %s", finalUserData.Name, finalUserData.Address, finalUserData.Role)
	} else {
		existingUser, ok := s.users[req.Name]
		if !ok {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("internal data inconsistency: user %s not found", req.Name))
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
	return c.Status(fiber.StatusCreated).JSON(LoginResponse{
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
func (s *Server) logoutHandler(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodPost {
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Invalid request method")
	}

	zlog.Info().Str("handler", "logoutHandler").Str("loggedInUser", s.loggedInUser).Msg("received logout request")

	if s.loggedInUser == "" {
		return c.Status(fiber.StatusBadRequest).SendString("No user is currently logged in")
	}

	loggedOutUser := s.loggedInUser
	s.loggedInUser = ""
	return c.JSON(fiber.Map{
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
func (s *Server) listUsersHandler(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodGet {
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Invalid request method")
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
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Failed to get address for key '%s': %v", record.Name, err))
		}

		pubKeyJSON, err := s.clientCtx.Codec.MarshalJSON(record.PubKey)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Failed to marshal pubkey for key '%s': %v", record.Name, err))
		}

		userInfos = append(userInfos, UserDetailResponse{
			Name:    record.Name,
			Type:    "local",
			Address: addr.String(),
			PubKey:  string(pubKeyJSON),
			Role:    userData.Role,
		})
	}

	return c.JSON(userInfos)
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
func (s *Server) requestKYCHandler(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodPost {
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Invalid request method")
	}
	if s.loggedInUser == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("No user is logged in. Please log in to request KYC.")
	}
	// Check if user already has a pending KYC request
	for _, req := range s.kycRequests {
		if req.Requester == s.loggedInUser && req.Status == "pending" {
			return c.Status(fiber.StatusBadRequest).SendString("You already have a pending KYC request.")
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
		zlog.Error().Err(err).Msg("failed to save KYC requests file")
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to save KYC request")
	}
	return c.Status(fiber.StatusCreated).JSON(newReq)
}

// getKYCRequestsHandler allows a regulator to view all pending KYC requests.
// @Summary Get pending KYC requests (regulator)
// @Description Allows a logged-in regulator to view all pending KYC requests.
// @Produce json
// @Success 200 {array} KYCRequestEntry
// @Failure 401 {object} KYCErrorResponse
// @Failure 403 {object} KYCErrorResponse
// @Router /user/kyc/requests [get]
func (s *Server) getKYCRequestsHandler(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodGet {
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Invalid request method")
	}
	if s.loggedInUser == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("No user is logged in.")
	}
	userData, ok := s.users[s.loggedInUser]
	if !ok || userData.Role != "regulator" {
		return c.Status(fiber.StatusForbidden).SendString("Only regulators can view KYC requests.")
	}
	pending := make([]KYCRequestEntry, 0)
	for _, req := range s.kycRequests {
		if req.Status == "pending" {
			pending = append(pending, req)
		}
	}
	return c.JSON(pending)
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
func (s *Server) approveKYCHandler(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodPost {
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Invalid request method")
	}
	if s.loggedInUser == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("No user is logged in.")
	}
	userData, ok := s.users[s.loggedInUser]
	if !ok || userData.Role != "regulator" {
		return c.Status(fiber.StatusForbidden).SendString("Only regulators can approve KYC requests.")
	}
	var req ApproveKYCRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
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
		return c.Status(fiber.StatusNotFound).SendString("KYC request not found or already processed")
	}
	s.saveKYCRequestsToFile()
	return c.JSON(KYCStatusResponse{Status: "approved"})
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
