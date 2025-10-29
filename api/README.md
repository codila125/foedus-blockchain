# API Module

RESTful HTTP server providing blockchain interaction endpoints for wallet management, smart contracts, and blockchain queries.

## Overview

The API module exposes the Foedus blockchain functionality via HTTP endpoints. Built with Chi router, it handles wallet operations, UTXO-based balance queries, and milestone-based smart contract lifecycle management with P2P network integration.

## Architecture

```
api/
├── routes.go          # Endpoint registration and server initialization
├── handler/           # HTTP request handlers with validation
│   └── handler.go
└── server/            # Core business logic and blockchain integration
    ├── server.go      # Blockchain operations
    ├── types.go       # Request/response structures
    └── response.go    # JSON serialization utilities
```

**Key Components:**
- **Handler**: Processes HTTP requests, validates addresses, returns JSON responses
- **Server**: Interfaces with blockchain, wallet, and P2P network layers
- **Response Transformers**: Convert binary blockchain data to JSON-safe formats

## API Endpoints

### Wallet Management

| Endpoint | Method | Description | Response |
|----------|--------|-------------|----------|
| `/blockchain/createwallet` | GET | Generate new wallet | `{"address": "..."}` (201) |
| `/blockchain/listaddresses` | GET | List all wallet addresses | `{"addresses": [...]}` (200) |
| `/blockchain/getbalance/{address}` | GET | Get UTXO balance for address | `{"balance": 100}` (200) |

### Blockchain Queries

| Endpoint | Method | Description | Response |
|----------|--------|-------------|----------|
| `/blockchain/printchain` | GET | Retrieve complete blockchain | Block array (200) |

### Smart Contract Operations

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/blockchain/createcontract/{address}` | POST | Create milestone-based contract |
| `/blockchain/getcontract/{contractID}` | GET | Retrieve contract details |
| `/blockchain/approvecontract/{address}/{contractID}` | GET | Sign and approve contract |
| `/blockchain/approvemilestone/{contractID}/{milestoneID}/{address}` | POST | Approve milestone with evidence |

## Usage Example

```bash
# Set node port and start server
export NODE_ID=3000
./blockchain

# Create wallet
curl http://localhost:3000/blockchain/createwallet

# Check balance
curl http://localhost:3000/blockchain/getbalance/{address}

# Create contract
curl -X POST http://localhost:3000/blockchain/createcontract/{creator_address} \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Project Alpha",
    "description": "Development work",
    "milestones": [{"title": "Phase 1", "value": 5000, "due_date": 1735689600}],
    "parties": [{"address": "{contractor_address}", "role": "CONTRACTOR"}],
    "terms": "terms_hash"
  }'
```

## Request/Response Schema

**Create Contract Request:**
```json
{
  "title": "string",
  "description": "string",
  "terms": "string",
  "milestones": [{
    "title": "string",
    "description": "string",
    "value": 1000,
    "due_date": 1735689600
  }],
  "parties": [{
    "address": "wallet_address",
    "role": "CONTRACTOR|CLIENT"
  }],
  "attachments": ["hash1"]
}
```

**Contract Response:**
```json
{
  "id": "hex_string",
  "title": "string",
  "creator": "address",
  "status": "PENDING|ACTIVE|COMPLETED",
  "created_at": "2025-01-01T00:00:00Z",
  "milestones": [{
    "id": "hex_string",
    "title": "string",
    "value": 1000,
    "status": "ACTIVE|APPROVED|COMPLETED",
    "approved_by": ["address1"]
  }],
  "parties": [{
    "address": "address",
    "role": "CONTRACTOR",
    "public_key": "hex_string"
  }]
}
```

## Error Handling

- **400** - Invalid wallet address format or malformed JSON
- **500** - Internal blockchain, wallet, or database errors

## Dependencies

- `github.com/go-chi/chi/v5` - HTTP router
- `github.com/libp2p/go-libp2p` - P2P networking
