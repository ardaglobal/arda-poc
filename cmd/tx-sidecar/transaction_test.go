package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

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