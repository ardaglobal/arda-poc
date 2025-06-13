package main

import (
	"encoding/json"
	"os"

	fiber "github.com/gofiber/fiber/v2"

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
	ID          string `json:"id"`
	Address     string `json:"address"`
	Region      string `json:"region"`
	Value       uint64 `json:"value"`
	TotalShares uint64 `json:"total_shares"`
	Status      string `json:"status"` // "for_sale", "pending_regulator_approval", "registered"
	Developer   string `json:"developer"`
}

// OffPlanPurchaseRequest represents a user's request to purchase shares in an off plan property.
type OffPlanPurchaseRequest struct {
	ID         string  `json:"id"`
	PropertyID string  `json:"property_id"`
	User       string  `json:"user"`
	AmountUSD  uint64  `json:"amount_usd"`
	Percent    float64 `json:"percent"`
	Status     string  `json:"status"` // "accepted"
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
	AmountUSD  uint64 `json:"amount_usd"`
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
func (s *Server) registerPropertyHandler(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodPost {
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Invalid request method")
	}

	var req RegisterPropertyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
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

	return s.buildSignAndBroadcast(c, fromName, req.Gas, "register_property", msgBuilder)
}

// transferSharesHandler handles share transfer
// @Summary Transfer property shares
// @Description Submits a transaction to transfer property shares between one or more owners.
// @Accept json
// @Produce json
// @Param request body TransferSharesRequest true "transfer details"
// @Success 200 {object} map[string]string{tx_hash=string}
// @Router /property/transfer-shares [post]
func (s *Server) transferSharesHandler(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodPost {
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Invalid request method")
	}

	var req TransferSharesRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
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

	if err := s.buildSignAndBroadcast(c, fromName, req.Gas, "transfer_shares", msgBuilder); err != nil {
		return err
	}

	// After transfer, update for-sale listings
	s.updateForSalePropertiesOnTransfer(req.PropertyID, req.FromOwners, req.FromShares)
	return nil
}

// editPropertyMetadataHandler edits property metadata
// @Summary Edit property metadata
// @Description Updates the metadata for an existing property.
// @Accept json
// @Produce json
// @Param request body EditPropertyMetadataRequest true "metadata"
// @Success 200 {object} map[string]string{tx_hash=string}
// @Router /property/edit [post]
func (s *Server) editPropertyMetadataHandler(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodPost {
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Invalid request method")
	}

	var req EditPropertyMetadataRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
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

	return s.buildSignAndBroadcast(c, fromName, req.Gas, "edit_property_metadata", msgBuilder)
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
func (s *Server) listPropertyForSaleHandler(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodPost {
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Invalid request method")
	}
	var req ListPropertyForSaleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
	}
	if req.PropertyID == "" || req.Owner == "" || len(req.Shares) == 0 || req.Price == 0 {
		return c.Status(fiber.StatusBadRequest).SendString("property_id, owner, shares, and price are required")
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
	return c.Status(fiber.StatusCreated).JSON(listing)
}

// getPropertiesForSaleHandler returns all properties currently for sale.
// @Summary Get properties for sale
// @Description Returns all properties currently listed for sale.
// @Produce json
// @Success 200 {array} ForSaleProperty
// @Router /property/for-sale [get]
func (s *Server) getPropertiesForSaleHandler(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodGet {
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Invalid request method")
	}
	listings := make([]ForSaleProperty, 0)
	for _, l := range s.forSaleProperties {
		if l.Status == "listed" {
			listings = append(listings, l)
		}
	}
	return c.JSON(listings)
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

// getOffPlanPurchaseRequestsHandler returns all purchase requests for a given off plan property.
// @Summary Get off plan property purchase requests
// @Description Returns all purchase requests for a given off plan property.
// @Produce json
// @Param property_id query string true "Off plan property ID"
// @Success 200 {array} OffPlanPurchaseRequest
// @Router /property/offplan/purchase-requests [get]
func (s *Server) getOffPlanPurchaseRequestsHandler(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodGet {
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Invalid request method")
	}
	propertyID := c.Query("property_id")
	if propertyID == "" {
		return c.Status(fiber.StatusBadRequest).SendString("property_id is required")
	}
	requests := make([]OffPlanPurchaseRequest, 0)
	for _, req := range s.offPlanPurchaseRequests {
		if req.PropertyID == propertyID {
			requests = append(requests, req)
		}
	}
	return c.JSON(requests)
}

// Persistence helpers for off plan properties and purchase requests
func (s *Server) saveOffPlanPropertiesToFile() error {
	data, err := json.MarshalIndent(s.offPlanProperties, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.offPlanPropertiesFile, data, 0644)
}

func (s *Server) saveOffPlanPurchaseRequestsToFile() error {
	data, err := json.MarshalIndent(s.offPlanPurchaseRequests, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.offPlanPurchaseRequestsFile, data, 0644)
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
func (s *Server) postOffPlanPropertyHandler(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodPost {
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Invalid request method")
	}
	if s.loggedInUser == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("No user is logged in.")
	}
	userData, ok := s.users[s.loggedInUser]
	if !ok || userData.Role != "developer" {
		return c.Status(fiber.StatusForbidden).SendString("Only developers can submit off plan properties.")
	}
	var req OffPlanPropertyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
	}
	if req.Address == "" || req.Region == "" || req.Value == 0 || req.TotalShares == 0 {
		return c.Status(fiber.StatusBadRequest).SendString("address, region, value, and total_shares are required")
	}
	newProp := OffPlanProperty{
		ID:          uuid.New().String(),
		Address:     req.Address,
		Region:      req.Region,
		Value:       req.Value,
		TotalShares: req.TotalShares,
		Status:      "for_sale",
		Developer:   s.loggedInUser,
	}
	s.offPlanProperties = append(s.offPlanProperties, newProp)
	s.saveOffPlanPropertiesToFile()
	return c.Status(fiber.StatusCreated).JSON(newProp)
}

// postOffPlanPurchaseRequestHandler allows a user to submit a purchase request for an off plan property.
// @Summary Submit off plan property purchase request
// @Description User submits a request to purchase shares in an off plan property. Auto-accepted if not fully funded. Rejected if >100% funded. If 100% funded, property status is set to 'pending_regulator_approval'.
// @Accept json
// @Produce json
// @Param request body OffPlanPurchaseRequestPayload true "purchase request info"
// @Success 201 {object} OffPlanPurchaseRequest
// @Failure 400 {object} KYCErrorResponse
// @Router /property/offplan/purchase-request [post]
func (s *Server) postOffPlanPurchaseRequestHandler(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodPost {
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Invalid request method")
	}
	if s.loggedInUser == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("No user is logged in.")
	}
	var req OffPlanPurchaseRequestPayload
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
	}
	// Find the property
	var prop *OffPlanProperty
	for i := range s.offPlanProperties {
		if s.offPlanProperties[i].ID == req.PropertyID {
			prop = &s.offPlanProperties[i]
			break
		}
	}
	if prop == nil {
		return c.Status(fiber.StatusBadRequest).SendString("Off plan property not found")
	}
	if prop.Status != "for_sale" {
		return c.Status(fiber.StatusBadRequest).SendString("Property is not for sale")
	}
	// Calculate current funding
	totalUSD := uint64(0)
	for _, pr := range s.offPlanPurchaseRequests {
		if pr.PropertyID == req.PropertyID {
			totalUSD += pr.AmountUSD
		}
	}
	if totalUSD+req.AmountUSD > prop.Value {
		return c.Status(fiber.StatusBadRequest).SendString("Purchase would exceed 100% funding")
	}
	percent := float64(req.AmountUSD) / float64(prop.Value) * 100.0
	newReq := OffPlanPurchaseRequest{
		ID:         uuid.New().String(),
		PropertyID: req.PropertyID,
		User:       s.loggedInUser,
		AmountUSD:  req.AmountUSD,
		Percent:    percent,
		Status:     "accepted",
	}
	s.offPlanPurchaseRequests = append(s.offPlanPurchaseRequests, newReq)
	s.saveOffPlanPurchaseRequestsToFile()
	// Check if property is now fully funded
	totalUSD += req.AmountUSD
	if totalUSD == prop.Value {
		prop.Status = "pending_regulator_approval"
		s.saveOffPlanPropertiesToFile()
	}
	return c.Status(fiber.StatusCreated).JSON(newReq)
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
func (s *Server) approveOffPlanPropertyHandler(c *fiber.Ctx) error {
	if c.Method() != fiber.MethodPost {
		return c.Status(fiber.StatusMethodNotAllowed).SendString("Invalid request method")
	}
	if s.loggedInUser == "" {
		return c.Status(fiber.StatusUnauthorized).SendString("No user is logged in.")
	}
	userData, ok := s.users[s.loggedInUser]
	if !ok || userData.Role != "regulator" {
		return c.Status(fiber.StatusForbidden).SendString("Only regulators can approve off plan properties.")
	}
	var req ApproveOffPlanPropertyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
	}
	// Find the property
	var prop *OffPlanProperty
	for i := range s.offPlanProperties {
		if s.offPlanProperties[i].ID == req.PropertyID {
			prop = &s.offPlanProperties[i]
			break
		}
	}
	if prop == nil {
		return c.Status(fiber.StatusBadRequest).SendString("Off plan property not found")
	}
	if prop.Status != "pending_regulator_approval" {
		return c.Status(fiber.StatusBadRequest).SendString("Property is not pending regulator approval")
	}
	// Gather owners and shares from purchase requests
	owners := []string{}
	shares := []uint64{}
	for _, pr := range s.offPlanPurchaseRequests {
		if pr.PropertyID == prop.ID {
			owners = append(owners, pr.User)
			// Shares proportional to percent of total shares
			share := uint64(float64(prop.TotalShares) * pr.Percent / 100.0)
			if share == 0 {
				share = 1
			} // Ensure at least 1 share
			shares = append(shares, share)
		}
	}
	if len(owners) == 0 {
		return c.Status(fiber.StatusBadRequest).SendString("No purchase requests found for this property")
	}
	regReq := RegisterPropertyRequest{
		Address: prop.Address,
		Region:  prop.Region,
		Value:   prop.Value,
		Owners:  owners,
		Shares:  shares,
	}
	// Broadcast property registration
	if err := s.buildSignAndBroadcast(c, "ERES", "", "register_property", func(fromAddr string) sdk.Msg {
		return propertytypes.NewMsgRegisterProperty(
			fromAddr,
			regReq.Address,
			regReq.Region,
			regReq.Value,
			regReq.Owners,
			regReq.Shares,
		)
	}); err != nil {
		return err
	}
	prop.Status = "registered"
	s.saveOffPlanPropertiesToFile()
	return c.JSON(prop)
}
