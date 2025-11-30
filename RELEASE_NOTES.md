# Foedus Blockchain v0.3 — Release Notes

## What's New in v0.3

### ✨ New Features

#### **Contract Cancellation** 🚫
Securely cancel active contracts and automatically recover undistributed funds.

**Features:**
- Cancel contracts at any time with proper authorization
- Proper authorization checks for contract parties
- Network-wide propagation of cancellation events
- Immutable cancellation records on blockchain

**API Endpoint:**
```bash
POST /blockchain/cancelcontract
```

**Example Usage:**
```bash
curl -X POST http://localhost:3000/blockchain/cancelcontract \
  -H "Content-Type: application/json" \
  -d '{
    "contract_id": "contract_hash_123",
    "creator": "creator_address"
  }'
```

### 🔧 Improvements
- Optimized contract state transition handling
- Better error handling for contract cancellation scenarios
- Enhanced contract validation logic

### 🐛 Bug Fixes
- Fixed race conditions in contract state updates
- Corrected contract status verification
- Resolved edge cases in fund recovery mechanism

---

**Download:** [Docker Hub](https://hub.docker.com/r/codila125/foedus-blockchain) | **Source:** [GitHub](https://github.com/codila125/foedus-blockchain)