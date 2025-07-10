package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	propertytypes "github.com/ardaglobal/arda-poc/x/property/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	zlog "github.com/rs/zerolog/log"
)

// RegisterPropertyRequest defines the request body for registering a property.
type RegisterPropertyRequest struct {
	Address string   `json:"address"`
	Region  string   `json:"region"`
	Value   uint64   `json:"value"`
	Owners  []string `json:"owners"`
	Shares  []uint64 `json:"shares"`
	Gas     string   `json:"gas,omitempty"`
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

// EditPropertyMetadataRequest defines the request body for editing property metadata.
type EditPropertyMetadataRequest struct {
	PropertyID              string `json:"property_id"`
	PropertyName            string `json:"property_name"`
	PropertyType            string `json:"property_type"`
	ParcelNumber            string `json:"parcel_number"`
	Size                    string `json:"size"`
	ConstructionInformation string `json:"construction_information"`
	ZoningClassification    string `json:"zoning_classification"`
	OwnerInformation        string `json:"owner_information"`
	TenantID                string `json:"tenant_id"`
	UnitNumber              string `json:"unit_number"`
	Gas                     string `json:"gas,omitempty"`
}

// ListPropertyForSaleRequest defines the request body for listing a property for sale.
type ListPropertyForSaleRequest struct {
	PropertyID string   `json:"property_id"`
	Owner      string   `json:"owner"`
	Shares     []uint64 `json:"shares"`
	Price      uint64   `json:"price"`
}

// OffPlanProperty represents a property to be funded off plan.
type OffPlanProperty struct {
	ID               string                   `json:"id"`
	Address          string                   `json:"address"`
	Region           string                   `json:"region"`
	Value            uint64                   `json:"value"`
	TotalShares      uint64                   `json:"total_shares"`
	Status           string                   `json:"status"` // "for_sale", "pending_regulator_approval", "registered"
	Developer        string                   `json:"developer"`
	PurchaseRequests []OffPlanPurchaseRequest `json:"purchase_requests,omitempty"`
}

// OffPlanPurchaseRequest represents a user's request to purchase shares in an off plan property.
type OffPlanPurchaseRequest struct {
	ID         string `json:"id"`
	PropertyID string `json:"property_id"`
	User       string `json:"user"`
	Shares     uint64 `json:"shares"`
	Status     string `json:"status"` // "accepted"
}

// OffPlanPropertyRequest defines the request body for submitting a new off plan property.
type OffPlanPropertyRequest struct {
	Address     string `json:"address"`
	Region      string `json:"region"`
	Value       uint64 `json:"value"`
	TotalShares uint64 `json:"total_shares"`
}

// OffPlanPurchaseRequestPayload defines the request body for submitting a purchase request.
type OffPlanPurchaseRequestPayload struct {
	PropertyID string `json:"property_id"`
	Shares     uint64 `json:"shares"`
}

// ApproveOffPlanPropertyRequest defines the request body for approving an off plan property.
type ApproveOffPlanPropertyRequest struct {
	PropertyID string `json:"property_id"`
}

// registerPropertyHandler handles property registration
// @Summary Register a property
// @Description Submits a transaction to register a new property on the blockchain.
// @Accept json
// @Produce json
// @Param request body RegisterPropertyRequest true "property info"
// @Success 200 {object} map[string]string{tx_hash=string}
// @Router /property/register [post]
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

	zlog.Info().Str("handler", "registerPropertyHandler").Interface("request", req).Msg("received request")

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

	s.buildSignAndBroadcast(w, r, fromName, req.Gas, "register_property", msgBuilder)
}

// transferSharesHandler handles share transfer
// @Summary Transfer property shares
// @Description Submits a transaction to transfer property shares between one or more owners.
// @Accept json
// @Produce json
// @Param request body TransferSharesRequest true "transfer details"
// @Success 200 {object} map[string]string{tx_hash=string}
// @Router /property/transfer-shares [post]
func (s *Server) transferSharesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		zlog.Error().Str("method", r.Method).Msg("invalid request method")
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req TransferSharesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		zlog.Error().Err(err).Msg("failed to decode request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	zlog.Info().Str("handler", "transferSharesHandler").Interface("request", req).Msg("received request")

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

	s.buildSignAndBroadcast(w, r, fromName, req.Gas, "transfer_shares", msgBuilder)

	// After transfer, update for-sale listings
	s.updateForSalePropertiesOnTransfer(req.PropertyID, req.FromOwners, req.FromShares)
}

// editPropertyMetadataHandler edits property metadata
// @Summary Edit property metadata
// @Description Updates the metadata for an existing property.
// @Accept json
// @Produce json
// @Param request body EditPropertyMetadataRequest true "metadata"
// @Success 200 {object} map[string]string{tx_hash=string}
// @Router /property/edit [post]
func (s *Server) editPropertyMetadataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req EditPropertyMetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	zlog.Info().Str("handler", "editPropertyMetadataHandler").Interface("request", req).Msg("received request")

	fromName := "ERES"
	msgBuilder := func(fromAddr string) sdk.Msg {
		return propertytypes.NewMsgEditPropertyMetadata(
			fromAddr,
			req.PropertyID,
			req.PropertyName,
			req.PropertyType,
			req.ParcelNumber,
			req.Size,
			req.ConstructionInformation,
			req.ZoningClassification,
			req.OwnerInformation,
			req.TenantID,
			req.UnitNumber,
		)
	}

	s.buildSignAndBroadcast(w, r, fromName, req.Gas, "edit_property_metadata", msgBuilder)
}

// listPropertyForSaleHandler allows an owner to list a property for sale.
// @Summary List property for sale
// @Description Allows an owner to list their property (or shares) for sale.
// @Accept json
// @Produce json
// @Param request body ListPropertyForSaleRequest true "listing info"
// @Success 201 {object} ForSaleProperty
// @Failure 400 {object} KYCErrorResponse
// @Router /property/list-for-sale [post]
func (s *Server) listPropertyForSaleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	var req ListPropertyForSaleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.PropertyID == "" || req.Owner == "" || len(req.Shares) == 0 || req.Price == 0 {
		http.Error(w, "property_id, owner, shares, and price are required", http.StatusBadRequest)
		return
	}
	listing := ForSaleProperty{
		ID:         uuid.New().String(),
		PropertyID: req.PropertyID,
		Owner:      req.Owner,
		Shares:     req.Shares,
		Price:      req.Price,
		Status:     "listed",
	}
	s.forSaleProperties = append(s.forSaleProperties, listing)
	s.saveForSalePropertiesToFile()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(listing)
}

// getPropertiesForSaleHandler returns all properties currently for sale.
// @Summary Get properties for sale
// @Description Returns all properties currently listed for sale.
// @Produce json
// @Success 200 {array} ForSaleProperty
// @Router /property/for-sale [get]
func (s *Server) getPropertiesForSaleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	listings := make([]ForSaleProperty, 0)
	for _, l := range s.forSaleProperties {
		if l.Status == "listed" {
			listings = append(listings, l)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(listings)
}

// saveForSalePropertiesToFile persists the forSaleProperties slice.
func (s *Server) saveForSalePropertiesToFile() error {
	data, err := json.MarshalIndent(s.forSaleProperties, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.forSalePropertiesFile, data, 0644)
}

// updateForSalePropertiesOnTransfer updates/removes for-sale listings after a transfer.
func (s *Server) updateForSalePropertiesOnTransfer(propertyID string, fromOwners []string, fromShares []uint64) {
	updatedListings := make([]ForSaleProperty, 0, len(s.forSaleProperties))

	for _, listing := range s.forSaleProperties {
		// Only update listings for the affected property and owner
		if listing.PropertyID == propertyID {
			for i, owner := range fromOwners {
				if listing.Owner == owner {
					// Update shares slice by subtracting the transferred shares
					if len(listing.Shares) == len(fromShares) {
						newShares := make([]uint64, len(listing.Shares))
						allZero := true
						for j := range listing.Shares {
							if i == j {
								if listing.Shares[j] > fromShares[j] {
									newShares[j] = listing.Shares[j] - fromShares[j]
									if newShares[j] > 0 {
										allZero = false
									}
								} else {
									newShares[j] = 0
								}
							} else {
								newShares[j] = listing.Shares[j]
								if newShares[j] > 0 {
									allZero = false
								}
							}
						}
						if !allZero {
							listing.Shares = newShares
							updatedListings = append(updatedListings, listing)
						}
						// If allZero, do not add this listing (it is now 0 shares)
						goto nextListing
					}
				}
			}
		}
		// If not updated, keep the listing as is
		updatedListings = append(updatedListings, listing)
		continue

	nextListing:
		continue
	}

	s.forSaleProperties = updatedListings
	s.saveForSalePropertiesToFile()
}

// getOffPlanPropertiesHandler returns all off-plan properties.
// @Summary Get all off-plan properties
// @Description Returns a list of all off-plan properties, including their purchase requests.
// @Produce json
// @Success 200 {array} OffPlanProperty
// @Router /property/offplans [get]
func (s *Server) getOffPlanPropertiesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	zlog.Info().Str("handler", "getOffPlanPropertiesHandler").Msg("received request")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.offPlanProperties)
}

// Persistence helpers for off plan properties and purchase requests
func (s *Server) saveOffPlanPropertiesToFile() error {
	data, err := json.MarshalIndent(s.offPlanProperties, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.offPlanPropertiesFile, data, 0644)
}

// postOffPlanPropertyHandler allows a developer to submit a new off plan property.
// @Summary Submit off plan property
// @Description Developer submits a new off plan property to be funded. Status is set to 'for_sale'.
// @Accept json
// @Produce json
// @Param request body OffPlanPropertyRequest true "off plan property info"
// @Success 201 {object} OffPlanProperty
// @Failure 400 {object} KYCErrorResponse
// @Router /property/offplan [post]
func (s *Server) postOffPlanPropertyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	if s.loggedInUser == "" {
		http.Error(w, "No user is logged in.", http.StatusUnauthorized)
		return
	}
	userData, ok := s.users[s.loggedInUser]
	if !ok || userData.Role != "developer" {
		http.Error(w, "Only developers can submit off plan properties.", http.StatusForbidden)
		return
	}
	var req OffPlanPropertyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Address == "" || req.Region == "" || req.Value == 0 || req.TotalShares == 0 {
		http.Error(w, "address, region, value, and total_shares are required", http.StatusBadRequest)
		return
	}
	newProp := OffPlanProperty{
		ID:               uuid.New().String(),
		Address:          req.Address,
		Region:           req.Region,
		Value:            req.Value,
		TotalShares:      req.TotalShares,
		Status:           "for_sale",
		Developer:        s.loggedInUser,
		PurchaseRequests: []OffPlanPurchaseRequest{}, // Initialize with empty slice
	}
	s.offPlanProperties = append(s.offPlanProperties, newProp)
	s.saveOffPlanPropertiesToFile()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newProp)
}

// postOffPlanPurchaseRequestHandler allows a user to submit a purchase request for an off plan property.
// @Summary Submit off plan property purchase request
// @Description User submits a request to purchase shares in an off plan property. When all shares are purchased, property status is set to 'pending_regulator_approval'.
// @Accept json
// @Produce json
// @Param request body OffPlanPurchaseRequestPayload true "purchase request info"
// @Success 201 {object} OffPlanPurchaseRequest
// @Failure 400 {object} KYCErrorResponse
// @Router /property/offplan/purchase-request [post]
func (s *Server) postOffPlanPurchaseRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	if s.loggedInUser == "" {
		zlog.Error().Msg("no user is logged in")
		http.Error(w, "No user is logged in.", http.StatusUnauthorized)
		return
	}
	var req OffPlanPurchaseRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	zlog.Info().Str("handler", "postOffPlanPurchaseRequestHandler").Interface("request", req).Msg("received request")
	// Find the property
	var prop *OffPlanProperty
	for i := range s.offPlanProperties {
		if s.offPlanProperties[i].ID == req.PropertyID {
			prop = &s.offPlanProperties[i]
			break
		}
	}
	if prop == nil {
		zlog.Error().Str("property_id", req.PropertyID).Msg("off plan property not found")
		http.Error(w, "Off plan property not found", http.StatusBadRequest)
		return
	}
	if prop.Status != "for_sale" {
		zlog.Error().Str("property_id", req.PropertyID).Msg("property is not for sale")
		http.Error(w, "Property is not for sale", http.StatusBadRequest)
		return
	}
	// Calculate current purchased shares
	totalSharesPurchased := uint64(0)
	for _, pr := range prop.PurchaseRequests {
		totalSharesPurchased += pr.Shares
	}

	if totalSharesPurchased+req.Shares > prop.TotalShares {
		zlog.Error().Str("property_id", req.PropertyID).Msg("purchase would exceed total shares")
		http.Error(w, "Purchase would exceed total shares", http.StatusBadRequest)
		return
	}

	newReq := OffPlanPurchaseRequest{
		ID:         uuid.New().String(),
		PropertyID: req.PropertyID,
		User:       s.loggedInUser,
		Shares:     req.Shares,
		Status:     "accepted",
	}
	prop.PurchaseRequests = append(prop.PurchaseRequests, newReq)

	// Check if property is now fully funded
	if totalSharesPurchased+req.Shares == prop.TotalShares {
		prop.Status = "pending_regulator_approval"
	}
	s.saveOffPlanPropertiesToFile()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newReq)
}

// approveOffPlanPropertyHandler allows a regulator to approve a fully funded off plan property, registering it on-chain.
// @Summary Approve off plan property (regulator)
// @Description Regulator approves a fully funded off plan property, registering it on-chain with the owners from the purchase requests. Status is updated to 'registered'.
// @Accept json
// @Produce json
// @Param request body ApproveOffPlanPropertyRequest true "Off plan property ID"
// @Success 200 {object} OffPlanProperty
// @Failure 400 {object} KYCErrorResponse
// @Failure 401 {object} KYCErrorResponse
// @Router /property/offplan/approve [post]
func (s *Server) approveOffPlanPropertyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	if s.loggedInUser == "" {
		zlog.Error().Msg("no user is logged in")
		http.Error(w, "No user is logged in.", http.StatusUnauthorized)
		return
	}
	userData, ok := s.users[s.loggedInUser]
	if !ok || userData.Role != "regulator" {
		zlog.Error().Str("loggedInUser", s.loggedInUser).Msg("user is not a regulator")
		http.Error(w, "Only regulators can approve off plan properties.", http.StatusForbidden)
		return
	}
	var req ApproveOffPlanPropertyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	zlog.Info().Str("handler", "approveOffPlanPropertyHandler").Interface("request", req).Msg("received request")
	// Find the property
	var prop *OffPlanProperty
	for i := range s.offPlanProperties {
		if s.offPlanProperties[i].ID == req.PropertyID {
			prop = &s.offPlanProperties[i]
			break
		}
	}
	if prop == nil {
		zlog.Error().Str("property_id", req.PropertyID).Msg("off plan property not found")
		http.Error(w, "Off plan property not found", http.StatusBadRequest)
		return
	}
	if prop.Status != "pending_regulator_approval" {
		zlog.Error().Str("property_id", req.PropertyID).Msg("property is not pending regulator approval")
		http.Error(w, "Property is not pending regulator approval", http.StatusBadRequest)
		return
	}
	// Gather owners and shares from purchase requests
	owners := []string{}
	shares := []uint64{}
	for _, pr := range prop.PurchaseRequests {
		userData, ok := s.users[pr.User]
		if !ok {
			zlog.Error().Str("user", pr.User).Msg("user from purchase request not found")
			http.Error(w, "User from purchase request not found", http.StatusInternalServerError)
			return
		}
		owners = append(owners, userData.Address)
		shares = append(shares, pr.Shares)
	}
	if len(owners) == 0 {
		http.Error(w, "No purchase requests found for this property", http.StatusBadRequest)
		return
	}
	// Register property on-chain (simulate by calling registerPropertyHandler logic)
	r2 := &http.Request{Body: io.NopCloser(strings.NewReader("")), Method: http.MethodPost}
	regReq := RegisterPropertyRequest{
		Address: prop.Address,
		Region:  prop.Region,
		Value:   prop.Value,
		Owners:  owners,
		Shares:  shares,
	}
	regReqBody, _ := json.Marshal(regReq)
	r2.Body = io.NopCloser(strings.NewReader(string(regReqBody)))
	s.registerPropertyHandler(w, r2)
	// Update property status
	prop.Status = "registered"
	s.saveOffPlanPropertiesToFile()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prop)
}
