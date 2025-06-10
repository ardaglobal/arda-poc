# Transaction Sidecar Service

This directory contains a simple Go HTTP service that acts as a "sidecar" to the main `ard-pocd` node. Its purpose is to provide a simple, high-level API for creating users and submitting transactions without requiring the client (e.g., a web UI) to handle the complexities of transaction creation, signing, and broadcasting.

This is a **proof-of-concept** to demonstrate server-side transaction signing. For a production-ready application that handles user funds, signing should be done on the client-side using a wallet like Keplr, so that private keys never leave the user's machine.

## How it Works

The service exposes several endpoints for interacting with the `arda-poc` blockchain. It uses the Cosmos SDK's Go libraries to programmatically:
1.  Manage user registration and login via email. It maintains a `logins.json` to map emails to on-server keyring names, and a `users.json` to store created user account details (including mnemonics).
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

### `POST /login`

Handles user login, registration, and linking.

The login flow is as follows:
- **Login:** If a user with the given `email` exists, they are logged in. The `name` field is ignored.
- **Register:** If the `email` does not exist and a `name` is provided, a new user account and key are created with that name. The new email is then linked to the new user, and they are logged in.
- **Link:** If the `email` does not exist but a user with the given `name` *does* exist, the email is linked to the existing user account, and they are logged in.
- **Error:** If the `email` does not exist and no `name` is provided, the request will fail, prompting the user to provide a name.

**Request Body:**

```json
{
  "email": "user@example.com",
  "name": "user-name",
  "role": "investor"
}
```
* `name` (string, optional): Required when registering a new user or linking an email to an existing user for the first time.
* `role` (string, optional): The user's role. Defaults to `user`. Allowed values: `user`, `investor`, `developer`, `regulator`, `admin`.

**Example `curl` Request (Login or Register/Link):**

```bash
# Login for an existing user
curl -X POST -H "Content-Type: application/json" -d '{"email": "bob@example.com"}' http://localhost:8080/login

# Register a new user with a specific role
curl -X POST -H "Content-Type: application/json" -d '{"email": "alice@example.com", "name": "alice", "role": "investor"}' http://localhost:8080/login
```

**Success Response (Login):**
```json
{
    "status": "success",
    "message": "User alice logged in",
    "user": "alice",
    "role": "investor"
}
```

**Success Response (Register/Link):**
```json
{
    "status": "success",
    "message": "User alice created/linked and logged in",
    "user": "alice",
    "role": "investor"
}
```

*Note: When a new user is created, their details (including the mnemonic and role) are saved to the server's `users.json` file. The mnemonic is not returned in the API response for security reasons.*

### `POST /logout`

Logs out the currently authenticated user.

**Example `curl` Request:**

```bash
curl -X POST http://localhost:8080/logout
```

**Success Response:**

```json
{
    "status": "success",
    "message": "User alice logged out"
}
```

### `GET /users`

Lists all registered users and their key details.

**Example `curl` Request:**

```bash
curl http://localhost:8080/users
```

**Success Response:**

Returns a JSON array of user details.
```json
[
    {
        "name": "ERES",
        "address": "arda1qzy8mf8epnpaetctnnhr28vl5h3d34jma8ev5y",
        "role": "admin",
        "type": "local",
        "pubkey": "{\"@type\":\"/cosmos.crypto.secp256k1.PubKey\",\"key\":\"A0O5s8d...\"}"
    },
    {
        "name": "bob",
        "address": "arda13pc7nj66w7cqsgs6kcn8x6n8a3gz76df7e552x",
        "role": "investor",
        "type": "local",
        "pubkey": "{\"@type\":\"/cosmos.crypto.secp256k1.PubKey\",\"key\":\"A8E7b2c...\"}"
    }
]
``` 