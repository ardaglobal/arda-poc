# Docker Network Setup

This repository provides Dockerfiles and a compose file for running a small `arda-poc` network. The network consists of the `ERES` validator and three full nodes that connect to it.

## Building Images

```bash
docker compose build
```

## Starting the Network

```bash
docker compose up -d
```

This command launches the validator and three full nodes. Ports are mapped as follows:

- Validator RPC: `localhost:26657`
- Full node 1 RPC: `localhost:26658`
- Full node 2 RPC: `localhost:26659`
- Full node 3 RPC: `localhost:26660`

## Interacting With Nodes

To run CLI commands against a node, use `docker exec` to access the container. For example, to query the validator:

```bash
docker exec -it validator arda-pocd status
```

Each node stores its data in a Docker volume. The validator also shares its `genesis.json` via the `genesis` volume so full nodes can start with the same chain state.

Stop the network with:

```bash
docker compose down
```
