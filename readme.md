# arda
**arda** is a blockchain built using Cosmos SDK and Tendermint and created with [Ignite CLI](https://ignite.com/cli).

## Development

```
make dev
```

This runs `ignite chain serve` with the `config.yml` file.

You can use https://github.com/ardaglobal/dexplorer to view the blockchain on the arda branch.

### Setup Development Environment

Run the `scripts/setup_dev_env.sh` script to install Go, Docker, Ignite CLI and
other tooling required for development. The script ensures all dependencies are
present and then executes `make setup-dev` to install additional Go tools and
protoc generators.

```
./scripts/setup_dev_env.sh
```

### Functionality

[demo](https://www.loom.com/share/7af628783a494d54bee0d8c6c8091041?sid=1a920879-d080-43d2-aea2-b6cc9986c705)

#### Submit Hash
Currently the blockchain is setup to submit a hash and a signature. This hash can represent any off chain data.

By running the tests in ./scripts you can generate the command to submit a hash to the blockchain and run the mock global node to track the valid hashes being submitted.

The chain name is arda and the token is uarda.

It is set up to be a single validator chain with the name of the validator being ERES.

#### Register Property

The property module is used to register properties on the blockchain.

It is used to register properties and their owners and shares.

#### Transfer Property

The property module is used to transfer property shares between owners.

#### USDArda

The USDArda module is used to mint and burn USDArda tokens.

### TODO

- integrate Keplr locally
- USDArda Minting from property registration
- USDArda transfers
- Update global node to track all successful events
- Deploy - dockerize to run both blockchain and ui together
- lazy block production? lazy block time like rollkit to reduce empty block spam?

### Configure

Your blockchain in development can be configured with `config.yml`. To learn more, see the [Ignite CLI docs](https://docs.ignite.com).

### Web Frontend

Additionally, Ignite CLI offers both Vue and React options for frontend scaffolding:

For a Vue frontend, use: `ignite scaffold vue`
For a React frontend, use: `ignite scaffold react`
These commands can be run within your scaffolded blockchain project. 


For more information see the [monorepo for Ignite front-end development](https://github.com/ignite/web).

## API

The REST API endpoints exposed by the modules under `x/` are derived from the gRPC services defined in `proto/ardapoc`. They can be accessed with any standard HTTP client.

### x/property

- `GET /arda/property/params` - query property module parameters
- `GET /cosmonaut/arda/property/properties` - list all registered properties
- `GET /cosmonaut/arda/property/properties/{index}` - get a property by index
- `POST /cosmonaut/arda/property/register` - register a property
- `POST /cosmonaut/arda/property/transfer` - transfer property shares

### x/arda

- `GET /arda/arda/params` - query arda module parameters
- `GET /cosmonaut/arda/arda/submissions` - list all submissions
- `GET /cosmonaut/arda/arda/submissions/{id}` - get a submission by id
- `POST /cosmonaut/arda/arda/submit-hash` - submit a hash

### Example: How to Register a Property via API

Registering a property, or performing any transaction via the API, is a multi-step process because all transactions must be cryptographically signed before being broadcast. The CLI handles this automatically, but to do it manually with a tool like `curl`, you must generate the transaction, sign it, and then broadcast it.

Here is a complete walkthrough using the `ard-pocd` binary and `curl`.

**Step 1: Create an Unsigned Transaction**

First, generate the transaction data without signing it. This command builds the transaction message and saves it as a JSON file. We'll use the `alice` key to create the property and make `bob` an owner.

```bash
./build/ard-pocd tx property register-property "123 Main St" "usa" 1000000 --owners "bob" --shares "100" --from "alice" --chain-id ardapoc --generate-only > unsigned-tx.json
```

**Step 2: Get Account Details for Signing**

To sign a transaction offline, you need the creator's `account_number` and their current `sequence` number. You can query the blockchain for this information.

```bash
# Query the 'alice' account
./build/ard-pocd query auth account arda1kc4y86tu4u58gzx80qhjv4hgzjf0g38h46wqkw -o json
```

The output will look something like this. You need the `account_number` and `sequence` values. If `account_number` is missing, it means the number is `0` (common for genesis accounts).

```json
{
  "account": {
    "type": "/cosmos.auth.v1beta1.BaseAccount",
    "value": {
      "address": "arda1kc4y86tu4u58gzx80qhjv4hgzjf0g38h46wqkw",
      "account_number": "2",
      "sequence": "0"
    }
  }
}
```

**Step 3: Sign the Transaction**

Now, sign the transaction using the details from the previous step. This creates a cryptographic signature and bundles it with the transaction data into a new file.

```bash
./build/ard-pocd tx sign unsigned-tx.json --from alice --offline --chain-id ardapoc --account-number 2 --sequence 0 > signed-tx.json
```

**Step 4: Prepare the Final JSON Payload for `curl`**

The API expects the signed transaction data to be encoded and wrapped in a specific JSON structure.

First, encode the signed transaction into the correct format:

```bash
TX_BYTES=$(./build/ard-pocd tx encode signed-tx.json)
```

Then, create a `payload.json` file with the required structure:

```bash
echo "{\"tx_bytes\":\"$TX_BYTES\",\"mode\":\"BROADCAST_MODE_SYNC\"}" > payload.json
```

**Step 5: Broadcast the Transaction with `curl`**

Finally, post the `payload.json` to the API's broadcast endpoint.

```bash
curl -X POST -H "Content-Type: application/json" http://0.0.0.0:1317/cosmos/tx/v1beta1/txs --data @payload.json
```

A successful response with `"code": 0` in the `tx_response` object indicates the transaction was successfully broadcast to the network. If you receive an `account sequence mismatch` error, it means the transaction succeeded and you need to increment the `--sequence` number for the next transaction from that account.

## Release
To release a new version of your blockchain, create and push a new tag with `v` prefix. A new draft release with the configured targets will be created.

```