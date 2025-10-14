# go_blockchain

A compact, production-oriented blockchain implementation in Go. This repository contains a simple Proof-of-Work blockchain, a UTXO model, wallet utilities, and a small P2P network layer for synchronizing blocks and transactions.

## Quick facts

- Language: Go 1.25+
- Storage: BadgerDB
- Consensus: Proof-of-Work (PoW)
- Transaction model: UTXO

## Installation

1. Install Go 1.25 or newer.
2. Clone the repository and download dependencies:

```bash
git clone https://github.com/yourusername/go_blockchain.git
cd go_blockchain
go mod download
go build ./...
```

## Basic usage

Set a NODE_ID environment variable (used to store node data under `temp/blocks_<NODE_ID>`):

```bash
export NODE_ID=3000
```

Create a wallet, then create a blockchain that funds the provided address with the genesis reward:

```bash
./blockchain createwallet
./blockchain createblockchain -address <ADDRESS>
```

Start a node (optionally with mining enabled):

```bash
export NODE_ID=3000
./blockchain startnode -miner <MINER_ADDRESS>
```

Send a transaction (mine locally with `-mine` or broadcast to the network):

```bash
./blockchain send -from <FROM> -to <TO> -amount <AMOUNT> [-mine]
```

Inspect the chain and UTXO set:

```bash
./blockchain printchain
./blockchain reindexutxo
```

## Project layout

- `blockchain/` – core blockchain code (blocks, transactions, UTXO set, PoW)
- `wallet/` – key management and address utilities
- `network/` – simple P2P handlers and message types
- `cli/` – command line interface and helpers
- `temp/` – local node data (per NODE_ID)

## Logging

The project uses structured, prefixed logs (e.g. `[BLOCKCHAIN]`, `[NETWORK]`, `[MINING]`). See `LOGGING_STANDARDS.md` for the logging conventions used across the codebase.

## Notes & future direction

- Current reward per mined block is hardcoded (see `blockchain/transaction.go`).
- No smart-contract engine yet; a planned future enhancement is milestone-based contract support and Wasm-based contracts.
- This code is a learning and experimental implementation—do not use it for production financial systems without further security review.

## License

MIT
# go_blockchain