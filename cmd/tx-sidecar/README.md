# Transaction Sidecar Service

This directory contains a simple Go HTTP service that acts as a "sidecar" to the main `ard-pocd` node. Its purpose is to provide a simple, high-level API for creating users and submitting transactions without requiring the client (e.g., a web UI) to handle the complexities of transaction creation, signing, and broadcasting.

This is a **proof-of-concept** to demonstrate server-side transaction signing. For a production-ready application that handles user funds, signing should be done on the client-side using a wallet like Keplr, so that private keys never leave the user's machine.

## How it Works

The service exposes several endpoints for interacting with the `arda-poc` blockchain. It uses the Cosmos SDK's Go libraries to programmatically:
1.  Manage user registration and login via email. It maintains a `logins.json` to map emails to on-server keyring names, and a `users.json` to store created user account details (including mnemonics).
2.  Load a pre-configured key (`ERES`) to sign administrative transactions like registering properties and transferring shares.
3.  Construct the appropriate messages (`MsgRegisterProperty`, `MsgTransferShares`, `MsgEditPropertyMetadata`).
4.  Query the blockchain for account details (number and sequence) needed for signing.
5.  Build, sign, encode, and broadcast the transaction to a local node via gRPC.
6.  Wait for the transaction to be included in a block and return the result.

## Running the Service

Ensure your `ard-pocd` node is running first. Then, from the root of the `arda-poc` repository, run the following command:

```bash
make dev-sidecar
```
This command uses [Air](https://github.com/cosmtrek/air) to watch the
`cmd/tx-sidecar` sources and automatically rebuild and restart the server when
files change. Make sure Air is installed:

```bash
go install github.com/cosmtrek/air@latest
```

Alternatively you can run:
```bash
go run ./cmd/tx-sidecar/main.go
```
The service will start and listen on port `8080`.

## API Documentation

This service uses Swagger to provide interactive API documentation. The documentation is automatically generated from the comments in the source code.

### Viewing the Docs

Once the sidecar service is running, you can view the API documentation by navigating to [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html) in your browser.

### Regenerating the Docs

To regenerate the OpenAPI specification files, run the following command from the root of the repository:

```bash
make sidecar-docs
```

This will update the auto-generated files in the `cmd/tx-sidecar/docs` directory based on the latest source code comments.