# API Module

RESTful HTTP server for blockchain operations.

## Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check |
| `/blockchain/createwallet` | GET | Generate wallet |
| `/blockchain/listaddresses` | GET | List wallets |
| `/blockchain/getbalance/{address}` | GET | Get balance |
| `/blockchain/printchain` | GET | Export blockchain |
| `/blockchain/createcontract/{address}` | POST | Create contract |
| `/blockchain/getcontract/{contractID}` | GET | Get contract |
| `/blockchain/approvecontract` | POST | Approve contract |
| `/blockchain/approvemilestone` | POST | Complete milestone |
| `/blockchain/cancelcontract` | POST | Cancel contract |

## Quick Start

```bash
# Start server
export NODE_ID=3000
./foedus

# Test health
curl http://localhost:3000/health
```

## Structure

```
api/
├── routes.go      # Endpoint registration
├── handler/       # Request handlers
└── server/        # Business logic
```
