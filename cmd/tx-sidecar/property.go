package main

import (
	"encoding/json"
	"net/http"
	"os"

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

// registerPropertyHandler handles property registration
// @Summary Register a property
// @Description Submits a transaction to register a new property on the blockchain.
// @Accept json
// @Produce json
// @Param request body RegisterPropertyRequest true "property info"
// @Success 200 {object} map[string]string{tx_hash=string}
// @Router /register-property [post]
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
// @Router /transfer-shares [post]
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
// @Router /edit-property [post]
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
// @Router /list-property-for-sale [post]
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
// @Router /properties-for-sale [get]
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
	// TODO: Implement logic to match and update for-sale listings when shares are transferred.
}
