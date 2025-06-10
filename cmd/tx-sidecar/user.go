package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

// UserDetailResponse defines the structure for returning detailed user information.
type UserDetailResponse struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	Role    string `json:"role"`
	Type    string `json:"type"`
	PubKey  string `json:"pubkey"`
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
			log.Printf("User '%s' not found, creating...", name)
			_, err := s.createUser(name, role)
			if err != nil {
				return fmt.Errorf("failed to create initial user '%s': %w", name, err)
			}
			log.Printf("Successfully created initial user '%s'", name)
		} else {
			// User exists, check if the role is correct.
			if userData.Role != role {
				log.Printf("User '%s' has incorrect role '%s', updating to '%s'.", name, userData.Role, role)
				userData.Role = role
				s.users[name] = userData
				if err := s.saveUsersToFile(); err != nil {
					return fmt.Errorf("failed to save users file after updating %s's role: %w", name, err)
				}
				log.Printf("Successfully updated role for user '%s'", name)
			}
		}
	}

	return nil
} 