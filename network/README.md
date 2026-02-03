# Network Module

P2P networking layer using libp2p.

## Node Types

| Type | Port | Purpose |
|------|------|---------|
| Source | 3009 | Network anchor, accepts connections |
| Miner | 3010 | Mining node, connects to source |

## Features

- Automatic peer discovery
- Blockchain synchronization
- Block/contract propagation
- Graceful shutdown handling

## Usage

```go
// Start source node
sourceNode := network.RunSourceNode(chain, nodeID)

// Start miner node
network.RunMinerNode(port, sourceMultiaddr)
```

## Protocols

| Protocol | Purpose |
|----------|---------|
| `GET_VERSION` | Check chain height |
| `GET_BLOCKCHAIN` | Sync full chain |
| `GET_BLOCKS` | Request specific blocks |
| `NEW_BLOCK` | Broadcast mined block |
| `NEW_CONTRACT` | Propagate contract |

## Structure

```
network/
├── source.go   # Source node
├── miner.go    # Miner node
├── handler.go  # Protocol handlers
├── peer.go     # Connection events
└── lib.go      # Sync utilities
```
