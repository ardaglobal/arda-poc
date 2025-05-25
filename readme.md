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


#### Submit Hash
Currently the blockchain is setup to submit a hash and a signature. This hash can represent any off chain data.

By running the tests in ./scripts you can generate the command to submit a hash to the blockchain and run the mock global node to track the valid hashes being submitted.

The chain name is arda and the token is uarda.

It is set up to be a single validator chain with the name of the validator being ERES.

#### Register Property

The property module is used to register properties on the blockchain.

It is used to register properties and their owners and shares.

#### Transfer Property

WIP

### Configure

Your blockchain in development can be configured with `config.yml`. To learn more, see the [Ignite CLI docs](https://docs.ignite.com).

### Docker Network

Dockerfiles and a compose file are provided to run a small network locally. Build the images and start the network with:

```bash
docker compose build
docker compose up -d
```

See [docs/docker-network.md](docs/docker-network.md) for details on interacting with the nodes.

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

## Release
To release a new version of your blockchain, create and push a new tag with `v` prefix. A new draft release with the configured targets will be created.

```
git tag v0.1
git push origin v0.1
```

After a draft release is created, make your final changes from the release page and publish it.

### Install
To install the latest version of your blockchain node's binary, execute the following command on your machine:

```
curl https://get.ignite.com/username/arda@latest! | sudo bash
```
`username/arda` should match the `username` and `repo_name` of the Github repository to which the source code was pushed. Learn more about [the install process](https://github.com/allinbits/starport-installer).

## Learn more

- [Ignite CLI](https://ignite.com/cli)
- [Tutorials](https://docs.ignite.com/guide)
- [Ignite CLI docs](https://docs.ignite.com)
- [Cosmos SDK docs](https://docs.cosmos.network)
- [Developer Chat](https://discord.gg/ignite)
