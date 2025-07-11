# Add Mortgage Rejection Functionality with Rejection Reasons

## Overview
Currently, the mortgage system only supports approval of mortgage requests. Lenders can approve pending mortgage requests through the `/bank/mortgage/create` endpoint, but there is no mechanism to reject mortgage requests with specific reasons.

## Current System Analysis

### Current Flow
1. **Lendee requests mortgage**: `POST /bank/mortgage/request` → Creates `MortgageRequest` with status "pending"
2. **Lender approves mortgage**: `POST /bank/mortgage/create` → Creates blockchain `Mortgage` with status `APPROVED`, updates `MortgageRequest` to "completed"

### Current Status Systems
- **MortgageRequest** (sidecar): `"pending"` → `"completed"`
- **Mortgage** (blockchain): `REQUESTED`, `APPROVED`, `REJECTED`, `PAID`, `CANCELLED`

### Gap Identified
- No endpoint to reject mortgage requests
- No storage for rejection reasons
- The `REJECTED` status exists in the protobuf but is unused

## Requirements

### 1. Rejection Endpoint
Add a new endpoint: `POST /bank/mortgage/reject`

**Request Body:**
```json
{
  "id": "mortgage-request-uuid",
  "reason": "High credit risk"
}
```

**Response:**
```json
{
  "status": "rejected",
  "reason": "High credit risk"
}
```

### 2. Rejection Reasons
Support predefined rejection reasons:
- `"High credit risk"`
- `"Insufficient documentation"`
- `"Insufficient collateral"`
- `"Income verification failed"`
- `"Property valuation issues"`
- `"Regulatory compliance issues"`
- `"Other"` (with optional free text)

### 3. Data Structure Updates

#### Update MortgageRequest Structure
Add rejection reason field:
```go
type MortgageRequest struct {
    // ... existing fields ...
    Status         string    `json:"status"`    // "pending", "completed", "rejected"
    RejectionReason string   `json:"rejection_reason,omitempty"`
}
```

#### Update Status Handling
- `"pending"` → `"rejected"` (when rejected)
- `"pending"` → `"completed"` (when approved)

### 4. Authorization
- Only the assigned lender can reject their mortgage requests
- Must be logged in
- Same authorization pattern as approval

### 5. API Integration
- Add route in `main.go`: `app.Post("/bank/mortgage/reject", ...)`
- Add Swagger documentation
- Follow existing error handling patterns

### 6. Data Persistence
- Update `saveMortgageRequestsToFile()` to persist rejection reasons
- Ensure rejected requests are excluded from pending requests list

## Implementation Details

### Handler Function Structure
```go
func (s *Server) rejectMortgageHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Validate request method and authentication
    // 2. Parse request body with ID and reason
    // 3. Find mortgage request by ID
    // 4. Validate request is pending and user is the lender
    // 5. Update status to "rejected" and set reason
    // 6. Save to file
    // 7. Return success response
}
```

### Request/Response Types
```go
type RejectMortgageRequest struct {
    ID     string `json:"id"`
    Reason string `json:"reason"`
}

type RejectMortgageResponse struct {
    Status string `json:"status"`
    Reason string `json:"reason"`
}
```

## Testing Requirements
1. **Unit tests** for rejection handler
2. **Integration tests** for complete rejection flow
3. **Validation tests** for rejection reasons
4. **Authorization tests** ensuring only lenders can reject
5. **Edge case tests** for invalid requests

## UI/UX Considerations
- Add reject button to mortgage request interface
- Dropdown for rejection reasons
- Display rejection reason in request history
- Notifications for rejection events

## Documentation Updates
- Update API documentation
- Add rejection endpoint to Swagger
- Update user guides for lenders
- Add rejection reason glossary

## Related Files to Modify
- `cmd/tx-sidecar/bank.go` - Add rejection handler
- `cmd/tx-sidecar/main.go` - Add rejection route
- `cmd/tx-sidecar/docs/swagger.yaml` - Add API documentation
- Data persistence files - Update JSON structure

## Acceptance Criteria
- [ ] Lender can reject mortgage requests with reasons
- [ ] Rejected requests show proper status and reason
- [ ] Rejected requests are not shown in pending requests
- [ ] Only assigned lender can reject their requests
- [ ] Rejection reasons are validated against predefined list
- [ ] API properly documented with Swagger
- [ ] Data persistence works correctly
- [ ] Error handling matches existing patterns
- [ ] Tests cover all scenarios

## Priority
**High** - This is a critical missing feature for the mortgage approval workflow.

## Estimated Effort
**Medium** - Follows existing patterns but requires careful integration with current approval flow.