# Network Module

Peer-to-peer networking layer using libp2p for blockchain synchronization and block propagation.

## Overview

The network module implements P2P communication for the Foedus blockchain using libp2p. It manages node connections, blockchain synchronization, block broadcasting, and contract propagation across the network. Nodes can operate as source nodes (stable anchors) or miner nodes (mining-enabled participants).

## Architecture

```
network/
├── source.go     # Source node initialization and management
├── peer.go       # Peer connection event handling
├── miner.go      # Miner node setup and lifecycle
├── handler.go    # Network request handlers (protocols)
└── lib.go        # Core networking utilities and synchronization
```

**Components:**
- **Source Node**: Stable network anchor listening on port 8006
- **Miner Node**: Mining-enabled node listening on port 8007
- **PeerNotifee**: Event listener for peer connections/disconnections
- **Protocol Handlers**: Request/response handlers for blockchain data
- **Sync Functions**: Blockchain synchronization and version checking

## Node Types

### Source Node

**Purpose**: Stable network entry point for peer connections

```go
sourceNode := RunSourceNode(chain, nodeID)
// Listens on: /ip4/0.0.0.0/tcp/8006
```

**Features:**
- Accepts incoming peer connections
- Handles network requests (GET_BLOCKCHAIN, GET_VERSION, etc.)
- Automatically syncs with peers on connection
- Broadcasts new blocks and contracts

### Miner Node

**Purpose**: Mining-enabled node that connects to source node

```go
RunMinerNode(nodeID, sourceMultiaddress)
// Listens on: /ip4/0.0.0.0/tcp/8007
```

**Features:**
- Connects to source node via multiaddress
- Syncs blockchain on startup
- Handles network requests
- Mines new blocks
- Graceful shutdown on SIGINT/SIGTERM

## Protocol Handlers

**Protocol ID**: `/foedus/1.0.0`

### GET_BLOCKCHAIN

Request entire blockchain from peer.

```go
// Request
SendCommand(stream, "GET_BLOCKCHAIN")

// Response
// Sends all blocks from genesis to latest
```

**Use Case**: Initial sync for new nodes

### GET_BLOCKS

Request blocks after specific height.

```go
// Request
heightData := &protobuf.HeightData{Height: 100}
// Send heightData

// Response
// Sends blocks with height > 100
```

**Use Case**: Incremental sync for partially synced nodes

### GET_VERSION

Request peer's blockchain version info.

```go
// Request
SendCommand(stream, "GET_VERSION")

// Response
versionData := &protobuf.VersionData{
    Height:   bestHeight,
    NodeId:   nodeID,
    LastHash: lastHash,
}
```

**Use Case**: Compare chain heights to determine sync need

### NEW_BLOCK

Receive and process new mined block.

```go
// Broadcast
blockData := &protobuf.BlockData{
    Command: "NEW_BLOCK",
    Hash:    block.Hash,
    Height:  block.Height,
    Data:    serializedBlock,
}
```

**Use Case**: Propagate newly mined blocks to network

### NEW_CONTRACT

Receive and process new contract.

```go
// Broadcast
contractData := &protobuf.Contract{
    Id:          contract.ID,
    Title:       contract.Title,
    Milestones:  milestones,
    Parties:     parties,
    // ... other fields
}
```

**Use Case**: Propagate new contracts for mining

## Peer Event Handling

**PeerNotifee**: Monitors network events

```go
type PeerNotifee struct {
    host   host.Host
    chain  *blockchain.BlockChain
    nodeID string
}
```

**Events:**
- `Connected`: Peer joins network → trigger sync check
- `Disconnected`: Peer leaves network → log event
- `Listen`, `ListenClose`: No-op
- `OpenedStream`, `ClosedStream`: No-op

**Auto-Sync**: When peer connects, automatically check if sync needed after 2-second delay

## Synchronization

### Initial Sync (GetBlockchain)

```go
// Download entire blockchain from peer
err := GetBlockchain(minerNode, peerID, nodeID)

// Process:
// 1. Request blockchain via GET_BLOCKCHAIN
// 2. Receive blocks in chronological order
// 3. Store blocks in database with batch writer
// 4. Update last hash
```

### Version Comparison (RequestVersionFromPeers)

```go
versions := RequestVersionFromPeers(node, chain)

// Compare heights
for peerID, info := range versions {
    if info.BestHeight > localHeight {
        // Peer has longer chain
    }
}
```

### Incremental Sync (SyncToLatestBlockchain)

```go
// Sync with peer that has longer chain
SyncToLatestBlockchain(node, chain, nodeID)

// Process:
// 1. Request version from all peers
// 2. Find peer with longest chain
// 3. Request blocks after local height
// 4. Add received blocks to chain
```

## Block Broadcasting

**Broadcast New Block:**
```go
HandleSendNewBlockRequest(node, block)

// Sends block to all connected peers via NEW_BLOCK command
```

**Process:**
1. Iterate through connected peers
2. Create stream to each peer
3. Send NEW_BLOCK command
4. Send serialized block data
5. Close stream

## Message Format

**Length-Prefixed Protocol:**
```
[4 bytes: length][N bytes: protobuf message]
```

All messages use this format for reliable streaming over TCP.

**Commands**: Sent as plain strings (e.g., "GET_BLOCKCHAIN")

**Data**: Serialized using Protocol Buffers (protobuf)

## Performance Features

**Batch Writing**: Uses BatchWriter for initial sync (flushes every 100 blocks)

**Connection Management**:
- Checks peer connectivity before creating streams
- Skips disconnected peers in broadcasts
- Auto-sync on peer connection with 2s delay

**Graceful Shutdown**: 30-second timeout for clean node shutdown

## Dependencies

- `github.com/libp2p/go-libp2p` - P2P networking
- `github.com/multiformats/go-multiaddr` - Network address format
- `google.golang.org/protobuf` - Message serialization

## Related Modules

- **Blockchain** - Core operations and data structures
- **Database** - Persistent storage for blocks
- **Proto** - Protocol buffer definitions
