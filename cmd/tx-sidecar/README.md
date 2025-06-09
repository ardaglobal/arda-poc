# Transaction Sidecar Service

This directory contains a simple Go HTTP service that acts as a "sidecar" to the main `ard-pocd` node. Its purpose is to provide a simple, high-level API for creating users and submitting transactions without requiring the client (e.g., a web UI) to handle the complexities of transaction creation, signing, and broadcasting.

This is a **proof-of-concept** to demonstrate server-side transaction signing. For a production-ready application that handles user funds, signing should be done on the client-side using a wallet like Keplr, so that private keys never leave the user's machine.

## How it Works

The service exposes several endpoints for interacting with the `arda-poc` blockchain. It uses the Cosmos SDK's Go libraries to programmatically:
1.  Manage an on-server keyring for creating users and storing their keys in a local `users.json` file.
2.  Load a pre-configured key (`ERES`) to sign administrative transactions like registering properties and transferring shares.
3.  Construct the appropriate messages (`MsgRegisterProperty`, `MsgTransferShares`).
4.  Query the blockchain for account details (number and sequence) needed for signing.
5.  Build, sign, encode, and broadcast the transaction to a local node via gRPC.
6.  Wait for the transaction to be included in a block and return the result.

## Running the Service

Ensure your `ard-pocd` node is running first. Then, from the root of the `arda-poc` repository, run the following command:

```bash
make dev-sidecar
# Or alternatively:
# go run ./cmd/tx-sidecar/main.go
```
The service will start and listen on port `8080`.

## API Endpoints

### `POST /register-property`

Submits a transaction to register a new property on the blockchain.

**Request Body:**

```json
{
  "address": "123 Sidecar Lane",
  "region": "dev",
  "value": 500000,
  "owners": ["arda13pc7nj66w7cqsgs6kcn8x6n8a3gz76df7e552x"],
  "shares": [100],
  "gas": "auto"
}
```
*   `gas` (string, optional): The gas limit for the transaction. Can be a specific number (e.g., `"300000"`) or `"auto"` to use the sidecar's default.

**Example `curl` Request:**

```bash
curl -X POST -H "Content-Type: application/json" -d '{
  "address": "123 Sidecar Lane",
  "region": "dev",
  "value": 500000,
  "owners": ["arda13pc7nj66w7cqsgs6kcn8x6n8a3gz76df7e552x"],
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

### `POST /transfer-shares`

Submits a transaction to transfer property shares between one or more owners.

**Request Body:**

```json
{
  "property_id": "123 main st, anytown, usa",
  "from_owners": ["arda1qzy8mf8epnpaetctnnhr28vl5h3d34jma8ev5y"],
  "from_shares": [25],
  "to_owners": ["arda1un6a2k876hqwhe75zv0k369yqmwexfj6qkuzsk"],
  "to_shares": [25],
  "gas": "auto"
}
```
*   `gas` (string, optional): The gas limit for the transaction. Can be a specific number (e.g., `"300000"`) or `"auto"` to use the sidecar's default.

**Example `curl` Request:**

```bash
curl -X POST http://localhost:8080/transfer-shares -H "Content-Type: application/json" -d '{
  "property_id": "123 main st, anytown, usa",
  "from_owners": ["arda13pc7nj66w7cqsgs6kcn8x6n8a3gz76df7e552x"],
  "from_shares": [25],
  "to_owners": ["arda1szmz6ttcd2n85sfdlhat04m4443l99kfj2ju63"],
  "to_shares": [25]
}'
```

**Success Response:**

A successful broadcast will return a JSON object containing the transaction hash.

```json
{
  "tx_hash": "A1B2C3D4E5F6..."
}
``` 

### `POST /create-user`

Creates a new user account (key) in the sidecar's keyring and saves it to `users.json`.

**Request Body:**

```json
{
  "name": "new-user-name"
}
```

**Example `curl` Request:**

```bash
curl -X POST -H "Content-Type: application/json" -d '{"name": "bob"}' http://localhost:8080/create-user
```

**Success Response:**

A successful request will return a JSON object with the new user's details, including their mnemonic. **Store the mnemonic securely!**
```json
{
    "name": "bob",
    "address": "arda1...",
    "mnemonic": "word1 word2 ..."
}
```

### `GET /keys`

Lists all keys currently managed by the sidecar's keyring.

**Example `curl` Request:**

```bash
curl http://localhost:8080/keys
```

**Success Response:**

Returns a JSON array of key information.
```json
[
    {
        "name": "ERES",
        "type": "local",
        "address": "arda1qzy8mf8epnpaetctnnhr28vl5h3d34jma8ev5y",
        "pubkey": "{\"@type\":\"/cosmos.crypto.secp256k1.PubKey\",\"key\":\"A0O5s8d...\"}"
    },
    {
        "name": "bob",
        "type": "local",
        "address": "arda13pc7nj66w7cqsgs6kcn8x6n8a3gz76df7e552x",
        "pubkey": "{\"@type\":\"/cosmos.crypto.secp256k1.PubKey\",\"key\":\"A8E7b2c...\"}"
    }
]
``` 