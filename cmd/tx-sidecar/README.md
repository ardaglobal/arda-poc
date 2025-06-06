# Transaction Sidecar Service

This directory contains a simple Go HTTP service that acts as a "sidecar" to the main `ard-pocd` node. Its purpose is to provide a simple, high-level API endpoint for submitting transactions without requiring the client (e.g., a web UI) to handle the complexities of transaction creation, signing, and broadcasting.

This is a **proof-of-concept** to demonstrate server-side transaction signing. For a production-ready application that handles user funds, signing should be done on the client-side using a wallet like Keplr, so that private keys never leave the user's machine.

## How it Works

The service exposes a single endpoint that takes property details as a simple JSON payload. It then uses the Cosmos SDK's Go libraries to programmatically:
1.  Load a specific, pre-configured key from the server's keyring (in this case, the `alice` key).
2.  Construct a `MsgRegisterProperty` message.
3.  Query the blockchain for the account's current number and sequence.
4.  Build, sign, and encode the transaction.
5.  Broadcast the signed transaction to the local node via gRPC.
6.  Return the resulting transaction hash to the caller.

## Running the Service

Ensure your `ard-pocd` node is running first. Then, from the root of the `arda-poc` repository, run the following command:

```bash
go run ./cmd/tx-sidecar/main.go
```
The service will start and listen on port `8080`.

## API Endpoint

### `POST /register-property`

Submits a transaction to register a new property on the blockchain.

**Request Body:**

```json
{
  "address": "123 Sidecar Lane",
  "region": "dev",
  "value": 500000,
  "owners": ["arda1shr8vdu7exvwdcwaptc9mq293d8m6vp53qpuh8"],
  "shares": [100]
}
```

**Example `curl` Request:**

```bash
curl -X POST -H "Content-Type: application/json" -d '{
  "address": "123 Sidecar Lane",
  "region": "dev",
  "value": 500000,
  "owners": ["arda1shr8vdu7exvwdcwaptc9mq293d8m6vp53qpuh8"],
  "shares": [100]
}' http://localhost:8080/register-property
```

**Success Response:**

A successful broadcast will return a JSON object containing the transaction hash.

```json
{
  "tx_hash": "A1B2C3D4E5F6..."
}
``` 