# 🌐 API Module

> Modern RESTful HTTP interface for seamless blockchain integration.

The API module exposes all blockchain operations through clean, well-documented REST endpoints. Built with Go Chi router for performance and simplicity, perfect for web apps, mobile clients, and integrations.

---

## ✨ Features

- 🔌 **RESTful Design** - Industry-standard HTTP endpoints
- 📦 **JSON Responses** - Easy integration with any client
- ⚡ **Chi Router** - High-performance request routing
- 🛡️ **Error Handling** - Meaningful error messages and status codes
- 📝 **Bruno Collection** - Pre-built API testing suite
- 🌐 **CORS Ready** - Cross-origin support for web apps

## 🏗️ Architecture

### Core Components

- **`server/`** - Server logic and business operations
  - `server.go` - Core server structure and methods
  - `types.go` - Request/response type definitions
  - `response.go` - Response formatting utilities

- **`handler/`** - HTTP request handlers
  - `handler.go` - Handler implementations for all endpoints

- **`routes.go`** - Route registration and configuration

### Key Structures

```go
type Server struct {
    NodeID string
}

type Handler struct {
    ctx    context.Context
    server *Server
}
```

## 🚀 Usage

### Starting the API Server

```bash
# Set your node ID (also used as port)
export NODE_ID=3000

# Run without CLI arguments to start API server
./blockchain

# Output:
# [SERVER] Starting server node in PORT:3000
```

The server will listen on `http://localhost:3000`

## 📡 API Endpoints

### Wallet Operations

#### Create Wallet
```bash
# Generate a new wallet address
curl http://localhost:3000/blockchain/createwallet

# Response:
{
  "address": "1A2B3C4D5E6F7G8H9I0J1K2L3M4N5O6P7Q8R9S"
}
```

#### List All Addresses
```bash
# Get all wallet addresses
curl http://localhost:3000/blockchain/listaddresses

# Response:
{
  "addresses": [
    "1A2B3C4D5E6F7G8H9I0J1K2L3M4N5O6P7Q8R9S",
    "1Z2Y3X4W5V6U7T8S9R0Q1P2O3N4M5L6K7J8I9H"
  ]
}
```

#### Get Balance
```bash
# Check balance for a specific address
curl http://localhost:3000/blockchain/getbalance/YOUR_ADDRESS

# Response:
{
  "balance": 150
}
```

### Blockchain Operations

#### Print Chain
```bash
# Retrieve all blocks in the blockchain
curl http://localhost:3000/blockchain/printchain

# Response: Array of blocks
[
  {
    "hash": "00001a2b3c4d5e...",
    "prevHash": "0000000000...",
    "height": 0,
    "timestamp": 1729684200,
    "transactions": [...],
    "nonce": 123456
  },
  {
    "hash": "00002b3c4d5e6f...",
    "prevHash": "00001a2b3c4d5e...",
    "height": 1,
    "timestamp": 1729684800,
    "transactions": [...],
    "nonce": 234567
  }
]
```

### Smart Contract Operations

#### Create Contract
```bash
# Create a new milestone-based contract
curl -X POST http://localhost:3000/blockchain/createcontract/CREATOR_ADDRESS \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Website Development Project",
    "description": "Build a full-stack e-commerce website",
    "terms": "QmXyz123...",
    "milestones": [
      {
        "title": "Design Phase",
        "description": "Complete UI/UX design mockups",
        "value": 1000,
        "dueDate": 1735689600
      },
      {
        "title": "Frontend Development",
        "description": "Implement responsive frontend",
        "value": 2000,
        "dueDate": 1738368000
      },
      {
        "title": "Backend & Deployment",
        "description": "API development and production deployment",
        "value": 2000,
        "dueDate": 1740960000
      }
    ],
    "parties": [
      {
        "address": "CONTRACTOR_ADDRESS",
        "role": "CONTRACTOR"
      },
      {
        "address": "ARBITRATOR_ADDRESS",
        "role": "ARBITRATOR"
      }
    ],
    "attachments": ["QmAbc456...", "QmDef789..."]
  }'

# Response:
{
  "contractID": "a1b2c3d4e5f6g7h8i9j0...",
  "status": "success",
  "message": "Contract created successfully"
}
```

#### Approve Contract
```bash
# Party signs and approves the contract
curl http://localhost:3000/blockchain/approvecontract/PARTY_ADDRESS/CONTRACT_ID

# Response:
{
  "status": "success",
  "message": "Contract approved and signed by PARTY_ADDRESS"
}
```

#### Get Contract
```bash
# Retrieve full contract details
curl http://localhost:3000/blockchain/getcontract/CONTRACT_ID

# Response:
{
  "id": "a1b2c3d4e5f6...",
  "title": "Website Development Project",
  "description": "Build a full-stack e-commerce website",
  "status": "ACTIVE",
  "createdAt": 1729684200,
  "updatedAt": 1729684200,
  "creatorAddress": "1A2B3C...",
  "milestones": [
    {
      "id": "m1a2b3c4...",
      "title": "Design Phase",
      "value": 1000,
      "status": "ACTIVE",
      "dueDate": 1735689600
    }
  ],
  "parties": [
    {
      "address": "1A2B3C...",
      "role": "CREATOR",
      "signature": "304502..."
    },
    {
      "address": "1Z2Y3X...",
      "role": "CONTRACTOR",
      "signature": "304502..."
    }
  ]
}
```

#### Approve Milestone
```bash
# Approve completion of a milestone
curl -X POST http://localhost:3000/blockchain/approvemilestone/CONTRACT_ID/MILESTONE_ID/APPROVER_ADDRESS

# Response:
{
  "status": "success",
  "message": "Milestone approved successfully"
}
```

## 🔗 Complete API Reference

| Method | Endpoint | Description | Request Body |
|--------|----------|-------------|--------------|
| GET | `/blockchain/createwallet` | Create new wallet | - |
| GET | `/blockchain/listaddresses` | List all addresses | - |
| GET | `/blockchain/getbalance/{address}` | Get balance | - |
| GET | `/blockchain/printchain` | Get all blocks | - |
| POST | `/blockchain/createcontract/{address}` | Create contract | Contract JSON |
| GET | `/blockchain/approvecontract/{address}/{contractID}` | Approve contract | - |
| GET | `/blockchain/getcontract/{contractID}` | Get contract | - |
| POST | `/blockchain/approvemilestone/{contractID}/{milestoneID}/{address}` | Approve milestone | - |

## 💡 Usage Examples

### Complete Workflow with API

#### 1. Create Two Wallets
```bash
# Creator wallet
CREATOR=$(curl -s http://localhost:3000/blockchain/createwallet | jq -r '.address')
echo "Creator: $CREATOR"

# Contractor wallet
CONTRACTOR=$(curl -s http://localhost:3000/blockchain/createwallet | jq -r '.address')
echo "Contractor: $CONTRACTOR"
```

#### 2. Check Initial Balances
```bash
curl http://localhost:3000/blockchain/getbalance/$CREATOR
curl http://localhost:3000/blockchain/getbalance/$CONTRACTOR
```

#### 3. Create a Contract
```bash
CONTRACT_ID=$(curl -s -X POST http://localhost:3000/blockchain/createcontract/$CREATOR \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Mobile App Development",
    "description": "iOS and Android app",
    "milestones": [
      {
        "title": "MVP Development",
        "description": "Core features",
        "value": 5000,
        "dueDate": 1735689600
      }
    ],
    "parties": [
      {
        "address": "'$CONTRACTOR'",
        "role": "CONTRACTOR"
      }
    ]
  }' | jq -r '.contractID')

echo "Contract ID: $CONTRACT_ID"
```

#### 4. Approve Contract (Both Parties)
```bash
# Creator approves
curl http://localhost:3000/blockchain/approvecontract/$CREATOR/$CONTRACT_ID

# Contractor approves
curl http://localhost:3000/blockchain/approvecontract/$CONTRACTOR/$CONTRACT_ID
```

#### 5. View Contract
```bash
curl http://localhost:3000/blockchain/getcontract/$CONTRACT_ID | jq
```

#### 6. Approve Milestone
```bash
# Get milestone ID from contract
MILESTONE_ID=$(curl -s http://localhost:3000/blockchain/getcontract/$CONTRACT_ID | jq -r '.milestones[0].id')

# Approve milestone
curl -X POST http://localhost:3000/blockchain/approvemilestone/$CONTRACT_ID/$MILESTONE_ID/$CREATOR
```

### Using with Different Ports

```bash
# Node 1
export NODE_ID=3000
./blockchain &

# Node 2
export NODE_ID=3001
./blockchain &

# Node 3
export NODE_ID=3002
./blockchain &

# Now you can interact with three different nodes
curl http://localhost:3000/blockchain/printchain
curl http://localhost:3001/blockchain/printchain
curl http://localhost:3002/blockchain/printchain
```

## 🔐 Request/Response Formats

### CreateContractReq
```json
{
  "title": "string",
  "description": "string",
  "terms": "string (hash)",
  "milestones": [
    {
      "title": "string",
      "description": "string",
      "value": number,
      "dueDate": number (unix timestamp)
    }
  ],
  "parties": [
    {
      "address": "string",
      "role": "CONTRACTOR|ARBITRATOR|CREATOR"
    }
  ],
  "attachments": ["string (hash)", ...]
}
```

### Standard Success Response
```json
{
  "status": "success",
  "message": "Operation completed",
  "data": {}
}
```

### Error Response
```json
{
  "error": "Error message description"
}
```

## 📁 File Structure

```
api/
├── routes.go                # Route registration
├── handler/
│   └── handler.go          # HTTP handlers
├── server/
│   ├── server.go           # Server logic
│   ├── types.go            # Request/Response types
│   └── response.go         # Response utilities
├── foedus-bruno-api/       # Bruno API collection
│   ├── bruno.json
│   ├── createwallet.bru
│   ├── listaddresses.bru
│   ├── getbalance.bru
│   ├── printchain.bru
│   ├── createcontract.bru
│   ├── approvecontract.bru
│   ├── getcontract.bru
│   └── approvemilestone.bru
└── README.md               # This file
```

## 🧪 Testing with Bruno

Bruno API collection is included in `foedus-bruno-api/` folder for easy testing:

1. Install [Bruno](https://www.usebruno.com/)
2. Open the `foedus-bruno-api` collection
3. Update environment variables if needed
4. Run requests sequentially

## ⚙️ Configuration

### Port Configuration
The API port is determined by the `NODE_ID` environment variable:
```bash
export NODE_ID=3000  # Server runs on :3000
export NODE_ID=8080  # Server runs on :8080
```

### CORS (if needed)
Add CORS middleware in `routes.go` for web applications.

## 🔗 Dependencies

- `github.com/go-chi/chi/v5` - HTTP router
- `encoding/json` - JSON handling
- `net/http` - HTTP server

## 📚 Related Modules

- **Handler** - HTTP request handlers
- **Server** - Business logic layer
- **Blockchain** - Core blockchain operations
- **Wallet** - Wallet management

## 🎯 Best Practices

- Always validate addresses before operations
- Use proper HTTP methods (GET for reads, POST for writes)
- Handle errors gracefully with appropriate status codes
- Set `NODE_ID` before starting the server
- Use JSON for all request/response bodies

---

*For more information, see the main [Foedus Blockchain README](../README.md)*
