# CLI Module

Command-line interface for blockchain operations.

## Commands

| Command | Description |
|---------|-------------|
| `createblockchain -address <addr>` | Initialize blockchain |
| `createwallet` | Generate new wallet |
| `listaddresses` | Show all addresses |
| `getbalance -address <addr>` | Check balance |
| `send -from <addr> -to <addr> -amount <n>` | Transfer tokens |
| `printchain` | Display blockchain |
| `reindex` | Rebuild UTXO index |
| `startnode -source <multiaddr>` | Start miner node |
| `--health` | Health check (for containers) |

## Usage

```bash
# Set node ID
export NODE_ID=3008

# Create blockchain
./foedus createblockchain -address <your-address>

# Check balance
./foedus getbalance -address <your-address>

# Health check (Docker)
./foedus --health
```

## Structure

```
cli/
├── cli.go       # Command parser
├── command.go   # Command implementations
└── format.go    # Output formatting
```
