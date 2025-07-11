package main

// Transaction Sidecar API docs
//
// @title Transaction Sidecar API
// @version 1.0
// @description Simple HTTP service for submitting blockchain transactions.
// @BasePath /

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	fiberadaptor "github.com/gofiber/adaptor/v2"
	fiber "github.com/gofiber/fiber/v2"
	fibercors "github.com/gofiber/fiber/v2/middleware/cors"
	fiberSwagger "github.com/gofiber/swagger"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/yaml.v3"

	"github.com/cosmos/cosmos-sdk/client"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	_ "github.com/ardaglobal/arda-poc/cmd/tx-sidecar/docs"
	sidecarclient "github.com/ardaglobal/arda-poc/pkg/client"
	"github.com/joho/godotenv"
)

// getEnv is a helper function to read an environment variable or return a fallback value.
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var (
	blockchainRestAPIURL = getEnv("BLOCKCHAIN_REST_API_URL", "http://localhost:1317")
	grpcAddr             = getEnv("GRPC_ADDR", "localhost:9090")
	nodeRPCURL           = getEnv("NODE_RPC_URL", "http://localhost:26657")
	faucetURL            = getEnv("FAUCET_URL", "http://localhost:4500")
)

func init() {
	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stdout})
}

// Config structs for parsing config.yml
type FaucetConfig struct {
	Name string `yaml:"name"`
}

type AppConfig struct {
	Faucet FaucetConfig `yaml:"faucet"`
}

// ForSaleProperty represents a property or shares listed for sale.
type ForSaleProperty struct {
	ID         string   `json:"id"` // unique listing ID
	PropertyID string   `json:"property_id"`
	Owner      string   `json:"owner"`
	Shares     []uint64 `json:"shares"`
	Price      uint64   `json:"price"`
	Status     string   `json:"status"` // "listed", "sold"
}

// Server holds the dependencies for the sidecar http server.
type Server struct {
	clientCtx            client.Context
	authClient           authtypes.QueryClient
	txClient             txtypes.ServiceClient
	users                map[string]UserData
	usersFile            string
	logins               map[string]string // email -> name
	loginsFile           string
	faucetName           string
	loggedInUser         string
	transactions         []TrackedTx
	transactionsFile     string
	mortgageRequests     []MortgageRequest
	mortgageRequestsFile string

	kycRequests     []KYCRequestEntry
	kycRequestsFile string

	forSaleProperties     []ForSaleProperty
	forSalePropertiesFile string

	offPlanProperties           []OffPlanProperty
	offPlanPropertiesFile       string
	offPlanPurchaseRequests     []OffPlanPurchaseRequest
	offPlanPurchaseRequestsFile string
}

// NewServer creates a new instance of the Server with all its dependencies.
func NewServer(clientCtx client.Context, grpcAddr string) (*Server, error) {
	grpcConn, err := grpc.NewClient(grpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server at %s: %w", grpcAddr, err)
	}

	dataDir := "cmd/tx-sidecar/local_data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	usersFile := filepath.Join(dataDir, "users.json")
	users := make(map[string]UserData)

	file, err := os.ReadFile(usersFile)
	if err == nil {
		if err := json.Unmarshal(file, &users); err != nil {
			zlog.Warn().Msgf("failed to unmarshal users file, starting with empty user map: %v", err)
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
				zlog.Warn().Msgf("failed to get address for key '%s', skipping sync: %v", record.Name, err)
				continue
			}
			users[record.Name] = UserData{
				Name:     record.Name,
				Address:  addr.String(),
				Mnemonic: "", // Mnemonic is not available from keyring listing
				Role:     "user",
			}
			zlog.Info().Msgf("Syncing key '%s' from keyring to users.json", record.Name)
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

	loginsFile := filepath.Join(dataDir, "logins.json")
	logins := make(map[string]string)

	loginData, err := os.ReadFile(loginsFile)
	if err == nil {
		if err := json.Unmarshal(loginData, &logins); err != nil {
			zlog.Warn().Msgf("failed to unmarshal logins file, starting with empty login map: %v", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read logins file: %w", err)
	}

	transactionsFile := filepath.Join(dataDir, "tx.json")
	transactions := make([]TrackedTx, 0)
	txData, err := os.ReadFile(transactionsFile)
	if err == nil {
		if err := json.Unmarshal(txData, &transactions); err != nil {
			zlog.Warn().Msgf("failed to unmarshal transactions file, starting with empty list: %v", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read transactions file: %w", err)
	}

	mortgageRequestsFile := filepath.Join(dataDir, "mortgage_requests.json")
	mortgageRequests := make([]MortgageRequest, 0)
	mrData, err := os.ReadFile(mortgageRequestsFile)
	if err == nil {
		if err := json.Unmarshal(mrData, &mortgageRequests); err != nil {
			zlog.Warn().Msgf("failed to unmarshal mortgage requests file, starting with empty list: %v", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read mortgage requests file: %w", err)
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

	kycRequestsFile := filepath.Join(dataDir, "kyc_requests.json")
	kycRequests := make([]KYCRequestEntry, 0)
	krData, err := os.ReadFile(kycRequestsFile)
	if err == nil {
		if err := json.Unmarshal(krData, &kycRequests); err != nil {
			zlog.Warn().Msgf("failed to unmarshal kyc requests file, starting with empty list: %v", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read kyc requests file: %w", err)
	}

	forSalePropertiesFile := filepath.Join(dataDir, "for_sale_properties.json")
	forSaleProperties := make([]ForSaleProperty, 0)
	fspData, err := os.ReadFile(forSalePropertiesFile)
	if err == nil {
		if err := json.Unmarshal(fspData, &forSaleProperties); err != nil {
			zlog.Warn().Msgf("failed to unmarshal for sale properties file, starting with empty list: %v", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read for sale properties file: %w", err)
	}

	offPlanPropertiesFile := filepath.Join(dataDir, "offplan_properties.json")
	offPlanProperties := make([]OffPlanProperty, 0)
	oppData, err := os.ReadFile(offPlanPropertiesFile)
	if err == nil {
		if err := json.Unmarshal(oppData, &offPlanProperties); err != nil {
			zlog.Warn().Msgf("failed to unmarshal off plan properties file, starting with empty list: %v", err)
		}
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to read off plan properties file: %w", err)
	}

	s := &Server{
		clientCtx:             clientCtx,
		authClient:            authtypes.NewQueryClient(grpcConn),
		txClient:              txtypes.NewServiceClient(grpcConn),
		users:                 users,
		usersFile:             usersFile,
		logins:                logins,
		loginsFile:            loginsFile,
		faucetName:            appConfig.Faucet.Name,
		transactions:          transactions,
		transactionsFile:      transactionsFile,
		mortgageRequests:      mortgageRequests,
		mortgageRequestsFile:  mortgageRequestsFile,
		kycRequests:           kycRequests,
		kycRequestsFile:       kycRequestsFile,
		forSaleProperties:     forSaleProperties,
		forSalePropertiesFile: forSalePropertiesFile,
		offPlanProperties:     offPlanProperties,
		offPlanPropertiesFile: offPlanPropertiesFile,
	}

	// Ensure that the faucet account from config exists in the keyring.
	if _, err := s.clientCtx.Keyring.Key(s.faucetName); err != nil {
		zlog.Warn().Msgf("faucet user '%s' from config.yml not found in keyring: %v", s.faucetName, err)
	} else {
		zlog.Info().Msgf("Faucet user '%s' found in keyring.", s.faucetName)
	}

	// Ensure faucet user has the 'bank' role.
	if faucetUserData, ok := s.users[s.faucetName]; ok {
		if faucetUserData.Role != "bank" {
			zlog.Info().Msgf("Updating role of bank user '%s' to 'bank'.", s.faucetName)
			faucetUserData.Role = "bank"
			s.users[s.faucetName] = faucetUserData
			if err := s.saveUsersToFile(); err != nil {
				zlog.Warn().Msgf("failed to save users file after updating bank role: %v", err)
			}
		}
	}

	if err := s.initUsers(); err != nil {
		return nil, fmt.Errorf("failed to initialize users: %w", err)
	}
	// Load KYC requests from file (in case not loaded above)
	_ = s.loadKYCRequestsFromFile()

	return s, nil
}

// Close is a no-op for this server version but can be used for cleanup.
func (s *Server) Close() {}

// AdminLoginRequest defines the request body for admin login.
type AdminLoginRequest struct {
	Key string `json:"key"`
}

// AdminLoginResponse defines the response for a successful admin login.
type AdminLoginResponse struct {
	Success bool `json:"success"`
}

// AdminLoginErrorResponse defines the response for an error in admin login.
type AdminLoginErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// adminLoginHandler handles admin login
// @Summary Admin login
// @Description Authenticates an admin using a key. Returns success if the provided key matches the ADMIN_KEY environment variable.
// @Accept json
// @Produce json
// @Param request body AdminLoginRequest true "Admin login key"
// @Success 200 {object} AdminLoginResponse
// @Failure 400 {object} AdminLoginErrorResponse
// @Failure 401 {object} AdminLoginErrorResponse
// @Failure 500 {object} AdminLoginErrorResponse
// @Router /admin/login [post]
func (s *Server) adminLoginHandler(c *fiber.Ctx) error {
	type reqBody AdminLoginRequest
	var body reqBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(AdminLoginErrorResponse{Success: false, Error: "invalid request body"})
	}
	adminKey := os.Getenv("ADMIN_KEY")
	if adminKey == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(AdminLoginErrorResponse{Success: false, Error: "admin key not set"})
	}
	if body.Key == adminKey {
		return c.JSON(AdminLoginResponse{Success: true})
	}
	return c.Status(fiber.StatusUnauthorized).JSON(AdminLoginErrorResponse{Success: false, Error: "invalid key"})
}

// passthroughGET proxies a GET request to the blockchain REST API and returns the response as-is.
func passthroughGET(path string, c *fiber.Ctx) error {
	baseURL := blockchainRestAPIURL
	// Compose the full URL
	url := baseURL + path
	resp, err := http.Get(url)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": err.Error()})
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	c.Status(resp.StatusCode)
	for k, v := range resp.Header {
		for _, vv := range v {
			c.Set(k, vv)
		}
	}
	return c.Send(body)
}

// getPropertiesPassthrough godoc
// @Summary      Proxy: Get all properties from blockchain
// @Description  Proxies GET /cosmonaut/arda/property/properties to the blockchain REST API
// @Tags         passthrough
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Failure      502 {object} map[string]string
// @Router       /cosmonaut/arda/property/properties [get]
func getPropertiesPassthrough(c *fiber.Ctx) error {
	return passthroughGET("/cosmonaut/arda/property/properties", c)
}

// getWalletBalancePassthrough godoc
// @Summary      Proxy: Get wallet balances
// @Description  Proxies GET /cosmos/bank/v1beta1/balances/{address} to the blockchain REST API
// @Tags         passthrough
// @Produce      json
// @Param        address path string true "Wallet address"
// @Success      200 {object} map[string]interface{}
// @Failure      502 {object} map[string]string
// @Router       /cosmos/bank/v1beta1/balances/{address} [get]
func getWalletBalancePassthrough(c *fiber.Ctx) error {
	address := c.Params("address")
	return passthroughGET("/cosmos/bank/v1beta1/balances/"+address, c)
}

// getMortgagesPassthrough godoc
// @Summary      Proxy: Get all mortgages from blockchain
// @Description  Proxies GET /ardaglobal/arda-poc/mortgage/mortgage to the blockchain REST API
// @Tags         passthrough
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Failure      502 {object} map[string]string
// @Router       /ardaglobal/arda-poc/mortgage/mortgage [get]
func getMortgagesPassthrough(c *fiber.Ctx) error {
	return passthroughGET("/ardaglobal/arda-poc/mortgage/mortgage", c)
}

func main() {
	// Load .env file if present
	_ = godotenv.Load()

	// This context is for the main application, not for individual requests.
	clientCtx, err := sidecarclient.NewClientContext()
	if err != nil {
		zlog.Fatal().Msgf("Failed to create client context: %v", err)
	}

	parsedURL, err := url.Parse(clientCtx.NodeURI)
	if err != nil {
		zlog.Fatal().Msgf("Failed to parse node URI: %v", err)
	}
	host := strings.Split(parsedURL.Host, ":")[0]
	clientCtxGrpcAddr := fmt.Sprintf("%s:9090", host)

	if clientCtxGrpcAddr != grpcAddr {
		zlog.Warn().Msgf("Using grpc addresses don't match: %s != %s", clientCtxGrpcAddr, grpcAddr)
	}
	// Use the client context's grpc address if it's set, otherwise use the default.
	server, err := NewServer(clientCtx, grpcAddr)
	if err != nil {
		zlog.Fatal().Msgf("Failed to create server: %v", err)
	}
	defer server.Close()

	developerUsers := make([]string, 0)
	investorUsers := make([]string, 0)
	for name, user := range server.users {
		switch user.Role {
		case "developer":
			developerUsers = append(developerUsers, name)
		case "investor":
			investorUsers = append(investorUsers, name)
		}
	}
	go server.RunAutoProperty(developerUsers, investorUsers)

	app := fiber.New()
	app.Use(fibercors.New())

	app.Get("/swagger/*", fiberSwagger.HandlerDefault)

	// User routes
	app.Get("/user/list", fiberadaptor.HTTPHandlerFunc(server.listUsersHandler))
	app.Post("/user/login", fiberadaptor.HTTPHandlerFunc(server.loginHandler))
	app.Post("/user/logout", fiberadaptor.HTTPHandlerFunc(server.logoutHandler))
	app.Get("/user/status", fiberadaptor.HTTPHandlerFunc(server.loginStatusHandler))

	// Property routes
	app.Post("/property/register", fiberadaptor.HTTPHandlerFunc(server.registerPropertyHandler))
	app.Post("/property/transfer-shares", fiberadaptor.HTTPHandlerFunc(server.transferSharesHandler))
	app.Post("/property/edit", fiberadaptor.HTTPHandlerFunc(server.editPropertyMetadataHandler))
	app.Post("/property/list-for-sale", fiberadaptor.HTTPHandlerFunc(server.listPropertyForSaleHandler))
	app.Get("/property/for-sale", fiberadaptor.HTTPHandlerFunc(server.getPropertiesForSaleHandler))

	// Off plan property routes
	app.Get("/property/offplans", fiberadaptor.HTTPHandlerFunc(server.getOffPlanPropertiesHandler))
	app.Post("/property/offplan", fiberadaptor.HTTPHandlerFunc(server.postOffPlanPropertyHandler))
	app.Post("/property/offplan/purchase-request", fiberadaptor.HTTPHandlerFunc(server.postOffPlanPurchaseRequestHandler))
	app.Post("/property/offplan/approve", fiberadaptor.HTTPHandlerFunc(server.approveOffPlanPropertyHandler))

	// Bank/mortgage routes
	app.Post("/bank/mortgage/request", fiberadaptor.HTTPHandlerFunc(server.requestMortgageHandler))
	app.Get("/bank/mortgage/requests", fiberadaptor.HTTPHandlerFunc(server.getMortgageRequestsHandler))
	app.Post("/bank/mortgage/create", fiberadaptor.HTTPHandlerFunc(server.createMortgageHandler))
	app.Post("/bank/mortgage/repay", fiberadaptor.HTTPHandlerFunc(server.repayMortgageHandler))
	app.Post("/bank/request-funds", fiberadaptor.HTTPHandlerFunc(server.requestFundsHandler))
	app.Post("/bank/mortgage/request-equity", fiberadaptor.HTTPHandlerFunc(server.requestEquityMortgageHandler))

	// KYC workflow
	app.Post("/user/kyc/request", fiberadaptor.HTTPHandlerFunc(server.requestKYCHandler))
	app.Get("/user/kyc/requests", fiberadaptor.HTTPHandlerFunc(server.getKYCRequestsHandler))
	app.Post("/user/kyc/approve", fiberadaptor.HTTPHandlerFunc(server.approveKYCHandler))

	// Transaction routes
	app.Get("/tx/list", fiberadaptor.HTTPHandlerFunc(server.listTransactionsHandler))
	app.Get("/tx/:hash", fiberadaptor.HTTPHandlerFunc(server.getTransactionHandler))

	// Admin
	app.Post("/admin/login", server.adminLoginHandler)

	// Passthrough routes for blockchain REST API (with Swagger docs)
	app.Get("/cosmonaut/arda/property/properties", getPropertiesPassthrough)
	app.Get("/cosmos/bank/v1beta1/balances/:address", getWalletBalancePassthrough)
	app.Get("/ardaglobal/arda-poc/mortgage/mortgage", getMortgagesPassthrough)

	zlog.Info().Msg("Starting transaction sidecar server on :8080...")
	if err := app.Listen(":8080"); err != nil {
		zlog.Fatal().Msgf("Failed to start server: %v", err)
	}
}
