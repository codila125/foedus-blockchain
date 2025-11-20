# Foedus Blockchain v0.2 — Release Notes

## What is Foedus?

**Foedus** is a secure, decentralized platform designed to revolutionize how teams collaborate on projects. Using cutting-edge blockchain technology, Foedus enables trustless agreements between multiple parties through **smart contracts**, ensuring that payment is released only when milestones are successfully completed.

Whether you're a freelancer, agency, or enterprise, Foedus provides transparency, security, and fairness in every contract.

---

## Key Highlights for v0.2

### 🚀 **Full-Featured Smart Contracts**
Create milestone-based project agreements with built-in multi-party verification. Define project phases, attach supporting documents, and automatically release payments when deliverables are approved.

### 💼 **Professional Project Management**
- Break projects into measurable milestones
- Attach documents and evidence to each phase
- Invite multiple parties (contractors, arbitrators, stakeholders)
- Track milestone completion with cryptographic proof

### 💰 **Transparent Payment System**
- Send and receive payments directly on the blockchain
- View your balance and transaction history in real-time
- No intermediaries, no hidden fees
- Transactions are immutable and auditable

### 🔒 **Bank-Grade Security**
- Industry-standard Ed25519 digital signatures
- SHA-256 cryptographic hashing
- Secure wallet management with Base58 addresses
- All transactions are tamper-proof and verifiable

### 🌐 **Decentralized Network**
Run your own node or connect to the Foedus peer-to-peer network. Foedus operates without central servers—you have complete control over your data and contracts.

### 📱 **Easy to Use**
Access Foedus via:
- **Simple Web API** — Integrate Foedus into your applications
- **Command-line Tools** — Perfect for developers and power users
- **Docker Containers** — One-click deployment, no setup required

---

## What You Can Do with Foedus

### **For Freelancers & Contractors**
- Bid on projects with confidence
- Get paid automatically when you deliver
- No disputes over incomplete work—milestones are verified on-chain

### **For Project Managers**
- Hire talent globally without intermediaries
- Track project progress through immutable milestones
- Release payments safely only when work is complete

### **For Enterprises**
- Build decentralized payment workflows
- Integrate blockchain verification into your systems
- Maintain full audit trails for compliance

### **For Developers**
- Deploy Foedus in seconds with Docker
- Build custom applications using the HTTP API
- Experiment with blockchain technology in a safe environment

---

## Getting Started

### **Option 1: Docker (Fastest)**
```bash
docker pull codila125/foedus-blockchain:latest
docker run -p 3000:3000 codila125/foedus-blockchain
```
Your blockchain is now running at `http://localhost:3000`

### **Option 2: From Source**
```bash
git clone https://github.com/codila125/foedus-blockchain.git
cd foedus-blockchain
go build -o foedus
./foedus
```

### **Step 1: Create a Wallet**
Your wallet is your identity on Foedus. It holds your tokens and signs your transactions.
```bash
curl http://localhost:3000/blockchain/createwallet
```

### **Step 2: Create a Contract**
Post a project with milestones and invite collaborators.
```bash
curl -X POST http://localhost:3000/blockchain/createcontract/{your_address} \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Logo Design",
    "description": "Design a professional company logo",
    "milestones": [
      {
        "title": "Concept Sketches",
        "value": 500,
        "due_date": 1762329600
      },
      {
        "title": "Final Designs",
        "value": 1500,
        "due_date": 1764921600
      }
    ],
    "parties": [{"address": "contractor_address", "role": "CONTRACTOR"}]
  }'
```

### **Step 3: Complete Milestones**
As work is completed, submit evidence. Once approved, payment is released automatically.

---

## Features in v0.2

✅ **Milestone-Based Smart Contracts** — Define and execute multi-phase projects  
✅ **Multi-Party Agreements** — Contracts involving creators, contractors, and arbitrators  
✅ **RESTful API** — 7 core endpoints for all blockchain operations  
✅ **Command-Line Interface** — Full-featured CLI for advanced users  
✅ **Peer-to-Peer Network** — Run multiple nodes for distributed consensus  
✅ **Docker Support** — Pre-configured images with automatic setup  
✅ **Persistent Storage** — Your blockchain data is safely stored locally  
✅ **Real-Time API Server** — Query balances, contracts, and blockchain state instantly  

---

## Security & Reliability

**Foedus is built on proven cryptographic standards:**
- **Consensus**: Proof-of-Work (same mining algorithm as Bitcoin)
- **Signatures**: Ed25519 (used by governments and enterprises worldwide)
- **Hashing**: SHA-256 (NSA-approved, industry standard)
- **Database**: Enterprise-grade PebbleDB with atomic transactions

Every transaction is immutable and verifiable. You can audit the entire blockchain history at any time.

---

## System Requirements

- **Memory**: 512MB minimum (1GB+ recommended)
- **Storage**: 200MB+ for blockchain data
- **Network**: Internet connection for P2P networking
- **Go Version**: 1.25.1+ (for building from source)
- **Docker**: Latest version (for containerized deployment)

---

## API Quick Reference

### Create a Wallet
```bash
GET /blockchain/createwallet
```

### Check Your Balance
```bash
GET /blockchain/getbalance/:address
```

### Send Tokens
```bash
POST /blockchain/send
```

### View Contract Status
```bash
GET /blockchain/getcontract/:contractID
```

### Approve a Milestone
```bash
POST /blockchain/approvemilestone
```

Full API documentation available in the project repository.

---

## Deployment Options

### **Local Development**
Perfect for testing and learning:
```bash
docker run -p 3000:3000 codila125/foedus-blockchain:latest
```

### **Production Network**
Run multiple nodes for true decentralization:
```bash
# Node 1 (Source)
NODE_ID=3000 ./foedus

# Node 2 (Miner)
NODE_ID=3001 ./foedus
```

### **Custom Configuration**
Override defaults with environment variables for your specific needs.

---

## Roadmap

🔮 **Coming Soon:**
- Enhanced contract templates for common use cases
- Mobile wallet applications
- Advanced analytics and reporting
- Cross-chain interoperability
- Layer 2 scaling for higher throughput

---

## Support & Community

- **Documentation**: See README.md for technical details
- **GitHub Issues**: Report bugs or request features
- **Docker Hub**: [codila125/foedus-blockchain](https://hub.docker.com/r/codila125/foedus-blockchain)
- **GitHub Container Registry**: ghcr.io/codila125/foedus-blockchain

---

## License

Foedus is open-source and released under the MIT License. You're free to use, modify, and distribute it.

---

## What's New in v0.2

### Improvements
- Optimized contract verification process
- Enhanced P2P network stability
- Improved Docker image efficiency

### Fixes
- Resolved edge cases in milestone approval workflow
- Fixed transaction serialization issues
- Corrected P2P peer discovery logic

---

**Start using Foedus today!**

[Deploy on Docker](https://hub.docker.com/r/codila125/foedus-blockchain) | [View Source Code](https://github.com/codila125/foedus-blockchain)

---