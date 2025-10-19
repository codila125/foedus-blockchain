package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/codila125/foedus-blockchain/wallet"
)

func (ct *Contract) SerializeContract() []byte {
	/*
		Serializes the contract into a byte array.
	*/
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(ct)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func CoinbaseOp(creator, data string) *Contract {
	/*
		Creates a coinbase contract that rewards the miner.
		'creator' is the address to send the reward to.
		'data' is arbitrary data, often used to include a message or extra information.
	*/
	if data == "" {
		randData := make([]byte, 20) // Generate 20 random bytes
		_, err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}
		data = fmt.Sprintf("%x", randData)
	}

	contract := &Contract{
		ID:             []byte{},
		Title:          "Coinbase Foedus",
		Description:    data,
		CreatedAt:      0,
		UpdatedAt:      0,
		Status:         ContractActive,
		Milestones:     []*Milestone{},
		Parties:        []*Party{},
		TermsHash:      []byte{},
		DisputeHandler: "",
		CreatorAddress: creator,
	}

	contract.ID = contract.HashContract()

	return contract
}

func (ct *Contract) IsCoinbaseOp() bool {
	/*
		Checks if the contract is a coinbase contract.
		Returns true if the contract has the title "Coinbase Foedus" and has no milestones or parties.
	*/
	return ct.Title == "Coinbase Foedus" && len(ct.Milestones) == 0 && len(ct.Parties) == 0
}

func (blockchain *BlockChain) FindContract(contractID string) (Contract, error) {
	/*
	   Finds a contract in the blockchain by its ID.
	   First checks the ICCT set for incomplete contracts (fast lookup),
	   then falls back to blockchain scan if not found.
	   contractID can be either hex string or raw bytes
	   Returns the contract and nil error if found, otherwise returns an error.
	*/

	// Convert hex string to bytes for comparison
	targetID, err := hex.DecodeString(contractID)
	if err != nil {
		// If it's not a valid hex string, treat it as raw bytes
		targetID = []byte(contractID)
	}

	// First, try to find in ICCT set (fast lookup for incomplete contracts)
	icctSet := ICCTSet{blockchain}
	contract, err := icctSet.GetContract(targetID)
	if err == nil && contract != nil {
		log.Printf("[CONTRACT] Contract %x found in ICCT set - Status: %s", targetID, contract.Status)
		return *contract, nil
	}

	// If not in ICCT, scan the entire blockchain (for completed/cancelled contracts)
	log.Printf("[CONTRACT] Contract %x not in ICCT set, scanning blockchain", targetID)
	iter := blockchain.Iterator()

	for {
		block := iter.Next()

		for _, ct := range block.Contracts {
			if bytes.Equal(ct.ID, targetID) {
				return *ct, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return Contract{}, errors.New("Contract not found")
}

func (ct *Contract) SignContract(privKey *ecdsa.PrivateKey, partyAddress []byte) error {
	/*
	   Signs a contract using the provided private key.
	   A party signs to indicate they agree to the contract terms.
	   The signature is appended to the party's entry in the Parties list.
	*/
	if ct.IsCoinbaseOp() {
		return nil
	}

	// Create a copy without existing signatures for hashing
	ctCopy := *ct

	// Deep copy the parties slice to avoid modifying the original
	ctCopy.Parties = make([]*Party, len(ct.Parties))
	for i := range ct.Parties {
		partyCopy := *ct.Parties[i]
		partyCopy.Signature = []byte{}
		ctCopy.Parties[i] = &partyCopy
	}

	// Hash the contract without signatures
	contractHash := ctCopy.HashContract()

	// Sign the contract hash
	r, s, err := ecdsa.Sign(rand.Reader, privKey, contractHash)
	if err != nil {
		return fmt.Errorf("failed to sign contract: %w", err)
	}

	// Encode signature as r||s (64 bytes total: 32 bytes r + 32 bytes s)
	signature := append(r.Bytes(), s.Bytes()...)

	// Find and update the party's signature
	found := false
	for i := range ct.Parties {
		if bytes.Equal([]byte(ct.Parties[i].Address), partyAddress) {
			ct.Parties[i].Signature = signature
			found = true
			log.Printf("[CONTRACT] ✓ Contract signed by party: %x", partyAddress)
			break
		}
	}

	if !found {
		return fmt.Errorf("party not found in contract: %x", partyAddress)
	}

	ct.UpdatedAt = time.Now().Unix()
	return nil
}

func (blockchain *BlockChain) VerifyContract(contract *Contract) bool {
	/*
	   Verifies the integrity of a contract.
	   DRAFT contracts can be added to blockchain (no signatures required yet).
	   ACTIVE contracts require all parties to have signed.
	   Returns true if the contract is valid for its current status.
	*/
	if contract.IsCoinbaseOp() {
		return true
	}

	// DRAFT contracts don't need signatures to be added to blockchain
	if contract.Status == ContractDraft {
		log.Printf("[CONTRACT] ✓ Contract %x verified - Status: DRAFT (awaiting signatures)", contract.ID)
		return true
	}

	// ACTIVE contracts must have all parties signed
	if contract.Status == ContractActive {
		if !contract.AreAllPartiesSigned() {
			unsigned := contract.GetUnsignedParties()
			log.Printf("[CONTRACT] ✗ Contract %x cannot be ACTIVE - %d parties still need to sign", contract.ID, len(unsigned))
			return false
		}

		// Verify each party's signature is valid
		for _, party := range contract.Parties {
			if !contract.VerifyContractSignature([]byte(party.Address), party.PublicKey) {
				log.Printf("[CONTRACT] ✗ Invalid signature from party: %s", party.Address)
				return false
			}
		}

		log.Printf("[CONTRACT] ✓ Contract %x verified - Status: ACTIVE (all %d parties signed)", contract.ID, len(contract.Parties))
		return true
	}

	return true
}

func (ct *Contract) VerifyContractSignature(partyAddress []byte, pubKeyBytes []byte) bool {
	/*
	   Verifies that a specific party has validly signed the contract.
	   Returns true if the signature is valid, false otherwise.
	*/
	if ct.IsCoinbaseOp() {
		return true
	}

	// Find the party and their signature
	var partySignature []byte
	for _, party := range ct.Parties {
		if bytes.Equal([]byte(party.Address), partyAddress) {
			partySignature = party.Signature
			break
		}
	}

	if len(partySignature) == 0 {
		log.Printf("[CONTRACT] ✗ No signature found for party: %x", partyAddress)
		return false
	}

	if len(pubKeyBytes) == 0 {
		log.Printf("[CONTRACT] ✗ No public key found for party: %x", partyAddress)
		return false
	}

	// Reconstruct the ECDSA public key from bytes
	// Public key format: X coordinate || Y coordinate (64 bytes total for P256)
	curve := elliptic.P256()
	keyLen := len(pubKeyBytes)

	if keyLen != 64 {
		log.Printf("[CONTRACT] ✗ Invalid public key length: %d (expected 64)", keyLen)
		return false
	}

	x := new(big.Int).SetBytes(pubKeyBytes[:32])
	y := new(big.Int).SetBytes(pubKeyBytes[32:])

	pubKey := &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}
	// Create contract copy without signatures for hashing
	ctCopy := *ct
	// Deep copy the parties slice to avoid modifying the original
	ctCopy.Parties = make([]*Party, len(ct.Parties))
	for i := range ct.Parties {
		partyCopy := *ct.Parties[i]
		partyCopy.Signature = []byte{}
		ctCopy.Parties[i] = &partyCopy
	}

	// Hash the contract
	contractHash := ctCopy.HashContract()

	// Decode signature (r||s format, 64 bytes)
	if len(partySignature) != 64 {
		log.Printf("[CONTRACT] ✗ Invalid signature length: %d (expected 64)", len(partySignature))
		return false
	}

	r := new(big.Int).SetBytes(partySignature[:32])
	s := new(big.Int).SetBytes(partySignature[32:])

	// Verify signature
	isValid := ecdsa.Verify(pubKey, contractHash, r, s)

	return isValid
}

func (ct *Contract) AreAllPartiesSigned() bool {
	/*
	   Checks if all parties have signed the contract.
	   Returns true only if every party has a non-empty signature.
	*/
	if ct.IsCoinbaseOp() {
		return true
	}

	for _, party := range ct.Parties {
		if len(party.Signature) == 0 {
			return false
		}
	}

	return true
}

func (ct *Contract) GetUnsignedParties() []*Party {
	/*
	   Returns a list of parties that have not yet signed the contract.
	*/
	var unsigned []*Party

	for _, party := range ct.Parties {
		if len(party.Signature) == 0 {
			unsigned = append(unsigned, party)
		}
	}

	return unsigned
}

func CreateContract(title, description string, w *wallet.Wallet, milestones []*Milestone, parties []*Party, termsHash []byte, disputeHandler string) *Contract {
	creatorAddress := string(w.Address())

	creatorParty := &Party{
		Address:   creatorAddress,
		Role:      RoleCreator,
		PublicKey: w.PublicKey,
		Signature: []byte{},
	}

	// Add creator to parties if not already present
	hasCreator := false
	for i := range parties {
		if strings.EqualFold(parties[i].Address, creatorAddress) {
			parties[i].PublicKey = w.PublicKey
			hasCreator = true
			break
		}
	}

	if !hasCreator {
		parties = append([]*Party{creatorParty}, parties...)
	}

	ct := Contract{
		Title:          title,
		Description:    description,
		CreatorAddress: string(creatorAddress),
		Milestones:     milestones,
		Parties:        parties,
		TermsHash:      termsHash,
		DisputeHandler: disputeHandler,
		CreatedAt:      time.Now().Unix(),
		UpdatedAt:      time.Now().Unix(),
		Status:         ContractDraft,
	}

	ct.ID = ct.HashContract()
	err := ct.ApproveContract(w)
	Handle(err)

	log.Printf("[CONTRACT] Contract created - ID: %x, Title: %s", ct.ID, title)

	return &ct
}

func (contract *Contract) ApproveContract(wallet *wallet.Wallet) error {
	/*
		Approves the contract by verifying the party's signature.
	*/
	if contract.IsCoinbaseOp() {
		return nil
	}

	// Check if this approval will complete all signatures
	willBeComplete := true
	for _, party := range contract.Parties {
		// Skip the current party (they're about to sign)
		if bytes.Equal([]byte(party.Address), wallet.Address()) {
			continue
		}
		// If any other party hasn't signed, it won't be complete
		if len(party.Signature) == 0 {
			willBeComplete = false
			break
		}
	}

	// Set status to ACTIVE if all parties will have signed after this approval
	if willBeComplete {
		contract.Status = ContractActive
		contract.UpdatedAt = time.Now().Unix()
	}

	privatekey, err := wallet.ReconstructECDSAKey() // Reconstruct the ECDSA private key from the wallet
	Handle(err)
	err = contract.SignContract(privatekey, wallet.Address())
	Handle(err)
	log.Printf("[CONTRACT] ✓ Party %x approved the contract %x", wallet.Address(), contract.ID)

	if contract.AreAllPartiesSigned() {
		contract.Status = ContractActive
		contract.UpdatedAt = time.Now().Unix()
		log.Printf("[CONTRACT] ✓ All parties have approved the contract %x", contract.ID)
	} else {
		unsigned := contract.GetUnsignedParties()
		log.Printf("[CONTRACT] ⧗ Contract %x still pending - %d more signatures needed", contract.ID, len(unsigned))
		for _, party := range unsigned {
			log.Printf("[CONTRACT]   - Waiting for: %s (%s)", party.Address, party.Role)
		}
	}

	return nil
}

func (contract Contract) String() string {
	/*
		Returns a string representation of the contract.
	*/
	var b strings.Builder
	b.WriteString(fmt.Sprintf("╭─── CONTRACT [%.16x...] ───╮\n", contract.ID))
	b.WriteString(fmt.Sprintf("│ Title: %s\n", contract.Title))
	b.WriteString(fmt.Sprintf("│ Status: %s\n", contract.Status))
	b.WriteString(fmt.Sprintf("│ Creator: %s\n", contract.CreatorAddress))

	b.WriteString(fmt.Sprintf("├─ PARTIES (%d)\n", len(contract.Parties)))
	if len(contract.Parties) == 0 {
		b.WriteString("│   (No Parties)\n")
	}
	for _, party := range contract.Parties {
		signed := "📝" // Signed
		if len(party.Signature) == 0 {
			signed = "⏳" // Pending
		}
		b.WriteString(fmt.Sprintf("│   %s [%s] %s\n", signed, party.Role, party.Address))
	}

	b.WriteString(fmt.Sprintf("├─ MILESTONES (%d)\n", len(contract.Milestones)))
	if len(contract.Milestones) == 0 {
		b.WriteString("│   (No Milestones)\n")
	}
	for i, milestone := range contract.Milestones {
		b.WriteString(fmt.Sprintf("│   [%d] %s (%s)\n", i+1, milestone.Title, milestone.Status))
		b.WriteString(fmt.Sprintf("│       Value: %d\n", milestone.Value))
	}

	b.WriteString("╰" + strings.Repeat("─", 50) + "╯")
	return b.String()
}
