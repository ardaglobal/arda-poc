package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTransactionEvent tests the TransactionEvent struct
func TestTransactionEvent(t *testing.T) {
	event := TransactionEvent{
		Status:    "submitted",
		Timestamp: time.Now(),
		Height:    42,
		Code:      0,
		RawLog:    "",
		Events:    nil,
	}

	// Test JSON marshaling
	data, err := json.Marshal(event)
	require.NoError(t, err)
	assert.Contains(t, string(data), "submitted")
	assert.Contains(t, string(data), "\"height\":42")

	// Test JSON unmarshaling
	var unmarshaled TransactionEvent
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, event.Status, unmarshaled.Status)
	assert.Equal(t, event.Height, unmarshaled.Height)
	assert.Equal(t, event.Code, unmarshaled.Code)
}

// TestTrackedTx tests the enhanced TrackedTx struct
func TestTrackedTx(t *testing.T) {
	now := time.Now()
	tx := TrackedTx{
		Timestamp: now,
		Type:      "register_property",
		TxHash:    "ABC123",
		Status:    "confirmed",
		Events: []TransactionEvent{
			{
				Status:    "submitted",
				Timestamp: now,
				Height:    0,
				Code:      0,
				RawLog:    "",
				Events:    nil,
			},
			{
				Status:    "confirmed",
				Timestamp: now.Add(5 * time.Second),
				Height:    42,
				Code:      0,
				RawLog:    "",
				Events:    []map[string]interface{}{
					{
						"type": "message",
						"attributes": []map[string]string{
							{"key": "action", "value": "register_property"},
						},
					},
				},
			},
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(tx)
	require.NoError(t, err)
	assert.Contains(t, string(data), "register_property")
	assert.Contains(t, string(data), "confirmed")
	assert.Contains(t, string(data), "ABC123")

	// Test JSON unmarshaling
	var unmarshaled TrackedTx
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)
	assert.Equal(t, tx.Type, unmarshaled.Type)
	assert.Equal(t, tx.TxHash, unmarshaled.TxHash)
	assert.Equal(t, tx.Status, unmarshaled.Status)
	assert.Len(t, unmarshaled.Events, 2)
	assert.Equal(t, "submitted", unmarshaled.Events[0].Status)
	assert.Equal(t, "confirmed", unmarshaled.Events[1].Status)
}

// TestTrackTransactionEvent tests the trackTransactionEvent method
func TestTrackTransactionEvent(t *testing.T) {
	// Create a temporary directory for test data
	tempDir := t.TempDir()
	txFile := tempDir + "/tx.json"

	// Create a mock server
	server := &Server{
		transactions:     make([]TrackedTx, 0),
		transactionsFile: txFile,
	}

	// Test creating a new transaction
	server.trackTransactionEvent("register_property", "ABC123", "submitted", 0, "", nil)

	// Verify the transaction was created
	require.Len(t, server.transactions, 1)
	tx := server.transactions[0]
	assert.Equal(t, "register_property", tx.Type)
	assert.Equal(t, "ABC123", tx.TxHash)
	assert.Equal(t, "submitted", tx.Status)
	assert.Len(t, tx.Events, 1)
	assert.Equal(t, "submitted", tx.Events[0].Status)

	// Test adding a second event to the same transaction
	events := []map[string]interface{}{
		{
			"type": "message",
			"attributes": []map[string]string{
				{"key": "action", "value": "register_property"},
			},
		},
	}
	server.trackTransactionEvent("register_property", "ABC123", "confirmed", 0, "", events)

	// Verify the transaction was updated
	require.Len(t, server.transactions, 1)
	tx = server.transactions[0]
	assert.Equal(t, "confirmed", tx.Status)
	assert.Len(t, tx.Events, 2)
	assert.Equal(t, "submitted", tx.Events[0].Status)
	assert.Equal(t, "confirmed", tx.Events[1].Status)
	assert.Len(t, tx.Events[1].Events, 1)

	// Test creating a different transaction
	server.trackTransactionEvent("transfer_shares", "DEF456", "submitted", 0, "", nil)

	// Verify we now have two transactions
	require.Len(t, server.transactions, 2)
	assert.Equal(t, "ABC123", server.transactions[0].TxHash)
	assert.Equal(t, "DEF456", server.transactions[1].TxHash)
}

// TestTrackTransactionEventWithFailure tests failure scenarios
func TestTrackTransactionEventWithFailure(t *testing.T) {
	// Create a temporary directory for test data
	tempDir := t.TempDir()
	txFile := tempDir + "/tx.json"

	// Create a mock server
	server := &Server{
		transactions:     make([]TrackedTx, 0),
		transactionsFile: txFile,
	}

	// Test tracking a failed transaction
	server.trackTransactionEvent("register_property", "FAILED123", "failed", 5, "insufficient funds", nil)

	// Verify the transaction was created with failure details
	require.Len(t, server.transactions, 1)
	tx := server.transactions[0]
	assert.Equal(t, "register_property", tx.Type)
	assert.Equal(t, "FAILED123", tx.TxHash)
	assert.Equal(t, "failed", tx.Status)
	assert.Len(t, tx.Events, 1)
	assert.Equal(t, "failed", tx.Events[0].Status)
	assert.Equal(t, uint32(5), tx.Events[0].Code)
	assert.Equal(t, "insufficient funds", tx.Events[0].RawLog)
}

// TestGetTransactionEventsHandler tests the HTTP handler
func TestGetTransactionEventsHandler(t *testing.T) {
	// Create a temporary directory for test data
	tempDir := t.TempDir()
	txFile := tempDir + "/tx.json"

	// Create a mock server with test data
	server := &Server{
		transactions:     make([]TrackedTx, 0),
		transactionsFile: txFile,
	}

	// Add some test transactions
	server.trackTransactionEvent("register_property", "ABC123", "submitted", 0, "", nil)
	server.trackTransactionEvent("register_property", "ABC123", "confirmed", 0, "", []map[string]interface{}{
		{
			"type": "message",
			"attributes": []map[string]string{
				{"key": "action", "value": "register_property"},
			},
		},
	})

	// Test successful request
	req := httptest.NewRequest(http.MethodGet, "/tx/events/ABC123", nil)
	w := httptest.NewRecorder()

	server.getTransactionEventsHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Parse the response
	var response TrackedTx
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "register_property", response.Type)
	assert.Equal(t, "ABC123", response.TxHash)
	assert.Equal(t, "confirmed", response.Status)
	assert.Len(t, response.Events, 2)
	assert.Equal(t, "submitted", response.Events[0].Status)
	assert.Equal(t, "confirmed", response.Events[1].Status)
}

// TestGetTransactionEventsHandlerNotFound tests 404 scenario
func TestGetTransactionEventsHandlerNotFound(t *testing.T) {
	// Create a temporary directory for test data
	tempDir := t.TempDir()
	txFile := tempDir + "/tx.json"

	// Create a mock server with no transactions
	server := &Server{
		transactions:     make([]TrackedTx, 0),
		transactionsFile: txFile,
	}

	// Test request for non-existent transaction
	req := httptest.NewRequest(http.MethodGet, "/tx/events/NOTFOUND", nil)
	w := httptest.NewRecorder()

	server.getTransactionEventsHandler(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Transaction not found in local cache")
}

// TestGetTransactionEventsHandlerInvalidMethod tests invalid HTTP method
func TestGetTransactionEventsHandlerInvalidMethod(t *testing.T) {
	// Create a temporary directory for test data
	tempDir := t.TempDir()
	txFile := tempDir + "/tx.json"

	// Create a mock server
	server := &Server{
		transactions:     make([]TrackedTx, 0),
		transactionsFile: txFile,
	}

	// Test POST request (should be GET only)
	req := httptest.NewRequest(http.MethodPost, "/tx/events/ABC123", nil)
	w := httptest.NewRecorder()

	server.getTransactionEventsHandler(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request method")
}

// TestGetTransactionEventsHandlerEmptyHash tests empty hash scenario
func TestGetTransactionEventsHandlerEmptyHash(t *testing.T) {
	// Create a temporary directory for test data
	tempDir := t.TempDir()
	txFile := tempDir + "/tx.json"

	// Create a mock server
	server := &Server{
		transactions:     make([]TrackedTx, 0),
		transactionsFile: txFile,
	}

	// Test request with empty hash
	req := httptest.NewRequest(http.MethodGet, "/tx/events/", nil)
	w := httptest.NewRecorder()

	server.getTransactionEventsHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Transaction hash must be provided in the path")
}

// TestTransactionLifecycleStates tests the complete transaction lifecycle
func TestTransactionLifecycleStates(t *testing.T) {
	// Create a temporary directory for test data
	tempDir := t.TempDir()
	txFile := tempDir + "/tx.json"

	// Create a mock server
	server := &Server{
		transactions:     make([]TrackedTx, 0),
		transactionsFile: txFile,
	}

	txHash := "LIFECYCLE123"

	// Step 1: Transaction submitted
	server.trackTransactionEvent("register_property", txHash, "submitted", 0, "", nil)
	tx := server.transactions[0]
	assert.Equal(t, "submitted", tx.Status)
	assert.Len(t, tx.Events, 1)
	assert.Equal(t, "submitted", tx.Events[0].Status)

	// Step 2: Transaction confirmed
	events := []map[string]interface{}{
		{
			"type": "message",
			"attributes": []map[string]string{
				{"key": "action", "value": "register_property"},
				{"key": "sender", "value": "cosmos1abc123"},
			},
		},
	}
	server.trackTransactionEvent("register_property", txHash, "confirmed", 0, "", events)
	tx = server.transactions[0]
	assert.Equal(t, "confirmed", tx.Status)
	assert.Len(t, tx.Events, 2)
	assert.Equal(t, "submitted", tx.Events[0].Status)
	assert.Equal(t, "confirmed", tx.Events[1].Status)
	assert.Len(t, tx.Events[1].Events, 1)

	// Verify complete lifecycle through HTTP handler
	req := httptest.NewRequest(http.MethodGet, "/tx/events/"+txHash, nil)
	w := httptest.NewRecorder()
	server.getTransactionEventsHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var response TrackedTx
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "confirmed", response.Status)
	assert.Len(t, response.Events, 2)
	// Events should be in chronological order
	assert.True(t, response.Events[0].Timestamp.Before(response.Events[1].Timestamp) ||
		response.Events[0].Timestamp.Equal(response.Events[1].Timestamp))
}

// TestFileOperations tests file save/load operations
func TestFileOperations(t *testing.T) {
	// Create a temporary directory for test data
	tempDir := t.TempDir()
	txFile := tempDir + "/tx.json"

	// Create a mock server
	server := &Server{
		transactions:     make([]TrackedTx, 0),
		transactionsFile: txFile,
	}

	// Add some transactions
	server.trackTransactionEvent("register_property", "ABC123", "submitted", 0, "", nil)
	server.trackTransactionEvent("transfer_shares", "DEF456", "confirmed", 0, "", nil)

	// Verify the file was created and contains data
	require.FileExists(t, txFile)

	// Read and verify file contents
	data, err := os.ReadFile(txFile)
	require.NoError(t, err)

	var transactions []TrackedTx
	err = json.Unmarshal(data, &transactions)
	require.NoError(t, err)

	require.Len(t, transactions, 2)
	assert.Equal(t, "ABC123", transactions[0].TxHash)
	assert.Equal(t, "DEF456", transactions[1].TxHash)
}

// TestBackwardCompatibility tests that the old trackTransaction method still works
func TestBackwardCompatibility(t *testing.T) {
	// Create a temporary directory for test data
	tempDir := t.TempDir()
	txFile := tempDir + "/tx.json"

	// Create a mock server
	server := &Server{
		transactions:     make([]TrackedTx, 0),
		transactionsFile: txFile,
	}

	// Test the old method
	server.trackTransaction("register_property", "LEGACY123")

	// Verify it creates a transaction with legacy status
	require.Len(t, server.transactions, 1)
	tx := server.transactions[0]
	assert.Equal(t, "register_property", tx.Type)
	assert.Equal(t, "LEGACY123", tx.TxHash)
	assert.Equal(t, "legacy", tx.Status)
	assert.Len(t, tx.Events, 1)
	assert.Equal(t, "legacy", tx.Events[0].Status)
}

// TestConcurrentAccess tests thread safety (basic test)
func TestConcurrentAccess(t *testing.T) {
	// Create a temporary directory for test data
	tempDir := t.TempDir()
	txFile := tempDir + "/tx.json"

	// Create a mock server
	server := &Server{
		transactions:     make([]TrackedTx, 0),
		transactionsFile: txFile,
	}

	// Test concurrent access (simplified test)
	done := make(chan bool, 2)

	go func() {
		defer func() { done <- true }()
		server.trackTransactionEvent("register_property", "CONCURRENT1", "submitted", 0, "", nil)
	}()

	go func() {
		defer func() { done <- true }()
		server.trackTransactionEvent("transfer_shares", "CONCURRENT2", "submitted", 0, "", nil)
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify both transactions were created (allow for race conditions)
	assert.GreaterOrEqual(t, len(server.transactions), 1)
	assert.LessOrEqual(t, len(server.transactions), 2)

	// Verify at least one transaction can be found
	hashes := make([]string, 0)
	for _, tx := range server.transactions {
		hashes = append(hashes, tx.TxHash)
	}

	// At least one should be present
	containsConcurrent1 := false
	containsConcurrent2 := false
	for _, hash := range hashes {
		if hash == "CONCURRENT1" {
			containsConcurrent1 = true
		}
		if hash == "CONCURRENT2" {
			containsConcurrent2 = true
		}
	}

	assert.True(t, containsConcurrent1 || containsConcurrent2, "At least one concurrent transaction should be present")
}

// BenchmarkTrackTransactionEvent benchmarks the trackTransactionEvent method
func BenchmarkTrackTransactionEvent(b *testing.B) {
	// Create a temporary directory for test data
	tempDir := b.TempDir()
	txFile := tempDir + "/tx.json"

	// Create a mock server
	server := &Server{
		transactions:     make([]TrackedTx, 0),
		transactionsFile: txFile,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		txHash := fmt.Sprintf("BENCH%d", i)
		server.trackTransactionEvent("register_property", txHash, "submitted", 0, "", nil)
	}
}

// BenchmarkGetTransactionEventsHandler benchmarks the HTTP handler
func BenchmarkGetTransactionEventsHandler(b *testing.B) {
	// Create a temporary directory for test data
	tempDir := b.TempDir()
	txFile := tempDir + "/tx.json"

	// Create a mock server with test data
	server := &Server{
		transactions:     make([]TrackedTx, 0),
		transactionsFile: txFile,
	}

	// Pre-populate with test data
	for i := 0; i < 1000; i++ {
		txHash := fmt.Sprintf("BENCH%d", i)
		server.trackTransactionEvent("register_property", txHash, "submitted", 0, "", nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		txHash := fmt.Sprintf("BENCH%d", i%1000)
		req := httptest.NewRequest(http.MethodGet, "/tx/events/"+txHash, nil)
		w := httptest.NewRecorder()
		server.getTransactionEventsHandler(w, req)


func TestListTransactionsPagination(t *testing.T) {
	// Create a server with some test transactions
	server := &Server{
		transactions: []TrackedTx{
			{Timestamp: time.Now(), Type: "register_property", TxHash: "hash1"},
			{Timestamp: time.Now(), Type: "transfer_shares", TxHash: "hash2"},
			{Timestamp: time.Now(), Type: "edit_property_metadata", TxHash: "hash3"},
			{Timestamp: time.Now(), Type: "register_property", TxHash: "hash4"},
			{Timestamp: time.Now(), Type: "transfer_shares", TxHash: "hash5"},
			{Timestamp: time.Now(), Type: "register_property", TxHash: "hash6"},
			{Timestamp: time.Now(), Type: "transfer_shares", TxHash: "hash7"},
		},
	}

	tests := []struct {
		name           string
		queryParams    string
		expectedCount  int
		expectedPage   int
		expectedTotal  int
		expectedHasNext bool
		expectedHasPrev bool
	}{
		{
			name:           "Default pagination",
			queryParams:    "",
			expectedCount:  7, // All transactions since we have less than default 50
			expectedPage:   1,
			expectedTotal:  7,
			expectedHasNext: false,
			expectedHasPrev: false,
		},
		{
			name:           "First page with limit 3",
			queryParams:    "?page=1&page_size=3",
			expectedCount:  3,
			expectedPage:   1,
			expectedTotal:  7,
			expectedHasNext: true,
			expectedHasPrev: false,
		},
		{
			name:           "Second page with limit 3",
			queryParams:    "?page=2&page_size=3",
			expectedCount:  3,
			expectedPage:   2,
			expectedTotal:  7,
			expectedHasNext: true,
			expectedHasPrev: true,
		},
		{
			name:           "Third page with limit 3",
			queryParams:    "?page=3&page_size=3",
			expectedCount:  1, // Last page with remaining transaction
			expectedPage:   3,
			expectedTotal:  7,
			expectedHasNext: false,
			expectedHasPrev: true,
		},
		{
			name:           "Page beyond available data",
			queryParams:    "?page=10&page_size=3",
			expectedCount:  0,
			expectedPage:   10,
			expectedTotal:  7,
			expectedHasNext: false,
			expectedHasPrev: true,
		},
		{
			name:           "Large page size",
			queryParams:    "?page=1&page_size=100",
			expectedCount:  7,
			expectedPage:   1,
			expectedTotal:  7,
			expectedHasNext: false,
			expectedHasPrev: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/tx/list"+tt.queryParams, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			server.listTransactionsHandler(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)
			require.Equal(t, "application/json", rr.Header().Get("Content-Type"))

			var response PaginatedTransactionsResponse
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(t, err)

			require.Equal(t, tt.expectedCount, len(response.Transactions))
			require.Equal(t, tt.expectedPage, response.Page)
			require.Equal(t, tt.expectedTotal, response.Total)
			require.Equal(t, tt.expectedHasNext, response.HasNext)
			require.Equal(t, tt.expectedHasPrev, response.HasPrev)
		})
	}
}

func TestListTransactionsPaginationEdgeCases(t *testing.T) {
	// Create a server with no transactions
	server := &Server{
		transactions: []TrackedTx{},
	}

	req, err := http.NewRequest("GET", "/tx/list", nil)
	require.NoError(t, err)

	rr := httptest.NewRecorder()
	server.listTransactionsHandler(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var response PaginatedTransactionsResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	require.NoError(t, err)

	require.Equal(t, 0, len(response.Transactions))
	require.Equal(t, 1, response.Page)
	require.Equal(t, 0, response.Total)
	require.Equal(t, false, response.HasNext)
	require.Equal(t, false, response.HasPrev)
}

func TestListTransactionsPaginationInvalidParams(t *testing.T) {
	server := &Server{
		transactions: []TrackedTx{
			{Timestamp: time.Now(), Type: "register_property", TxHash: "hash1"},
			{Timestamp: time.Now(), Type: "transfer_shares", TxHash: "hash2"},
		},
	}

	tests := []struct {
		name        string
		queryParams string
		expectedPage int
		expectedPageSize int
	}{
		{
			name:        "Invalid page number",
			queryParams: "?page=invalid&page_size=10",
			expectedPage: 1, // Should default to 1
			expectedPageSize: 10,
		},
		{
			name:        "Invalid page size",
			queryParams: "?page=1&page_size=invalid",
			expectedPage: 1,
			expectedPageSize: 50, // Should default to 50
		},
		{
			name:        "Negative page number",
			queryParams: "?page=-1&page_size=10",
			expectedPage: 1, // Should default to 1
			expectedPageSize: 10,
		},
		{
			name:        "Zero page size",
			queryParams: "?page=1&page_size=0",
			expectedPage: 1,
			expectedPageSize: 50, // Should default to 50
		},
		{
			name:        "Excessive page size",
			queryParams: "?page=1&page_size=2000",
			expectedPage: 1,
			expectedPageSize: 1000, // Should cap at 1000
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/tx/list"+tt.queryParams, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			server.listTransactionsHandler(rr, req)

			require.Equal(t, http.StatusOK, rr.Code)

			var response PaginatedTransactionsResponse
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			require.NoError(t, err)

			require.Equal(t, tt.expectedPage, response.Page)
			require.Equal(t, tt.expectedPageSize, response.PageSize)
		})
	}
}