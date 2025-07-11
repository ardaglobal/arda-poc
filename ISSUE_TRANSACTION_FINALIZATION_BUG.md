# Issue: Transaction Details Not Available Due to Incomplete Finalization

**Issue Type:** Bug  
**Priority:** Medium  
**Component:** tx-sidecar, autoproperty.go  
**Reported by:** Toby Thwaites  
**Date:** July 11, 2025  

## Problem Description

Users are encountering "No transaction details available" error when viewing transaction details in the frontend. This occurs intermittently when transactions are broadcast but not fully finalized within the polling timeout period.

## Root Cause Analysis

The issue is in the `buildSignAndBroadcastInternal` function in `cmd/tx-sidecar/transaction.go`:

1. **Transaction Broadcasting**: Uses `BROADCAST_MODE_SYNC` mode which returns immediately after mempool validation
2. **Polling Timeout**: Has a 30-second timeout to wait for block inclusion
3. **Tracking Logic**: Only adds transactions to the tracked list after successful block confirmation
4. **Frontend Dependency**: The UI queries transaction details from this tracked list

**The Problem Flow:**
1. Transaction is broadcast successfully (gets hash)
2. 30-second polling timeout expires before block inclusion
3. Transaction is NOT added to tracked transactions list
4. Frontend queries for transaction details but finds nothing in local cache
5. User sees "No transaction details available"

## Affected Code Areas

### Primary Location
- `cmd/tx-sidecar/transaction.go` - `buildSignAndBroadcastInternal()` function (lines 41-144)

### Secondary Locations  
- `cmd/tx-sidecar/autoproperty.go` - All calls to `buildSignAndBroadcastInternal()`:
  - Line 132: `registerProperty()`
  - Line 231: `transferShares()`  
  - Line 303: `autoEditPropertyMetadata()`

## Evidence

- **Screenshot**: Shows transaction hash with "No transaction details available" message
- **Slack Discussion**: Matt confirmed this is "an issue with the auto scripts" where "the transaction wasn't fully finalized"
- **Code Review**: 30-second timeout may be insufficient for block times during network congestion

## Proposed Solutions

### Option 1: Increase Polling Timeout
- Increase the 30-second timeout to 60-90 seconds
- **Pros**: Simple fix, handles most cases
- **Cons**: Doesn't solve the fundamental issue

### Option 2: Implement Graceful Degradation
- Track transactions immediately after broadcast (before block confirmation)
- Add a "pending" status for unconfirmed transactions
- Allow frontend to display pending transactions with appropriate status
- **Pros**: Better user experience, no data loss
- **Cons**: More complex implementation

### Option 3: Background Transaction Tracker
- Implement a background service that continues polling for transactions
- Remove the timeout constraint from the broadcast function
- **Pros**: Most robust solution
- **Cons**: Requires architectural changes

## Immediate Actions Required

1. **Short-term**: Increase polling timeout to 60 seconds
2. **Medium-term**: Implement graceful degradation (Option 2)
3. **Long-term**: Consider background transaction tracker (Option 3)

## Testing Strategy

1. **Reproduce Issue**: Create test scenario with slow block times
2. **Verify Fix**: Ensure transactions are properly tracked even with delays
3. **Load Testing**: Test with multiple concurrent autoproperty operations
4. **UI Testing**: Verify frontend properly displays transaction states

## Related Files

- `cmd/tx-sidecar/transaction.go` - Main transaction handling logic
- `cmd/tx-sidecar/autoproperty.go` - Auto property operations that trigger the issue
- `cmd/tx-sidecar/bank.go` - Additional transaction broadcasting calls

## Monitoring

Consider adding metrics for:
- Transaction broadcast success rate
- Average time to block inclusion
- Polling timeout occurrences
- Frontend "no details available" errors

---

**Assignee**: Chad  
**Labels**: bug, tx-sidecar, autoproperty, transaction-finalization  
**Milestone**: Next Release