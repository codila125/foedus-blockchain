package blockchain

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/codila125/foedus-blockchain/wallet"
)

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
		Terms:          []byte{},
		CreatorAddress: creator,
	}

	contract.ID = contract.HashContract()

	return contract
}

func (contract *Contract) IsCoinbaseOp() bool {
	/*
		Checks if the contract is a coinbase contract.
		Returns true if the contract has the title "Coinbase Foedus" and has no milestones or parties.
	*/
	return contract.Title == "Coinbase Foedus" && len(contract.Milestones) == 0 && len(contract.Parties) == 0
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

		for _, contract := range block.Contracts {
			if bytes.Equal(contract.ID, targetID) {
				return *contract, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return Contract{}, errors.New("Contract not found")
}

func (contract *Contract) SignContract(privKey *ed25519.PrivateKey, partyAddress []byte) error {
	/*
	   Signs a contract using the provided private key.
	   A party signs to indicate they agree to the contract terms.
	   The signature is appended to the party's entry in the Parties list.
	*/
	if contract.IsCoinbaseOp() {
		return nil
	}

	// Create a copy without existing signatures for hashing
	ctCopy := *contract
	ctCopy.Parties = make([]*Party, len(contract.Parties))
	for i, party := range contract.Parties {
		partyCopy := *party
		partyCopy.Signature = nil // Clear signature for hashing
		ctCopy.Parties[i] = &partyCopy
	}

	// Hash the contract without signatures
	contractHash := ctCopy.HashContract()

	// Sign the contract hash
	signature := ed25519.Sign(*privKey, contractHash)

	// Find and update the party's signature
	for i := range contract.Parties {
		if contract.Parties[i].Address == string(partyAddress) {
			contract.Parties[i].Signature = signature
			log.Printf("[CONTRACT] ✓ Contract signed by party: %s", contract.Parties[i].Address)
			contract.UpdatedAt = time.Now().Unix()
			return nil
		}
	}

	log.Printf("[CONTRACT] ✗ Party not found in contract: %s", string(partyAddress))
	return fmt.Errorf("[CONTRACT] ✗ Party not found in contract: %s", string(partyAddress))
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

func (contract *Contract) VerifyContractSignature(partyAddress []byte, pubKeyBytes []byte) bool {
	/*
	   Verifies that a specific party has validly signed the contract.
	   Returns true if the signature is valid, false otherwise.
	*/
	if contract.IsCoinbaseOp() {
		return true
	}

	// Find the party and their signature
	var partySignature []byte
	for _, party := range contract.Parties {
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

	// Create contract copy without signatures for hashing
	ctCopy := *contract
	// Deep copy the parties slice to avoid modifying the original
	ctCopy.Parties = make([]*Party, len(contract.Parties))
	for i, party := range contract.Parties {
		partyCopy := *party
		partyCopy.Signature = nil // Clear signature for hashing
		ctCopy.Parties[i] = &partyCopy
	}

	// Hash the contract
	contractHash := ctCopy.HashContract()

	// Decode signature (r||s format, 64 bytes)
	if len(partySignature) != 64 {
		log.Printf("[CONTRACT] ✗ Invalid signature length: %d (expected 64)", len(partySignature))
		return false
	}

	pubKey := ed25519.PublicKey(pubKeyBytes)

	return ed25519.Verify(pubKey, contractHash, partySignature)
}

func (contract *Contract) AreAllPartiesSigned() bool {
	/*
	   Checks if all parties have signed the contract.
	   Returns true only if every party has a non-empty signature.
	*/
	if contract.IsCoinbaseOp() {
		return true
	}

	for _, party := range contract.Parties {
		if len(party.Signature) == 0 {
			return false
		}
	}

	return true
}

func (contract *Contract) GetUnsignedParties() []*Party {
	/*
	   Returns a list of parties that have not yet signed the contract.
	*/
	unsigned := make([]*Party, 0)

	for _, party := range contract.Parties {
		if len(party.Signature) == 0 {
			unsigned = append(unsigned, party)
		}
	}

	return unsigned
}

func CreateContract(title, description string, w *wallet.Wallet, milestones []*Milestone, parties []*Party, terms []byte, attachments [][]byte) *Contract {
	creatorAddress := string(w.Address())

	creatorParty := &Party{
		Address:   creatorAddress,
		Role:      RoleCreator,
		PublicKey: []byte(w.PublicKey),
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

	contract := Contract{
		Title:          title,
		Description:    description,
		CreatorAddress: string(creatorAddress),
		Milestones:     milestones,
		Parties:        parties,
		Terms:          terms,
		CreatedAt:      time.Now().Unix(),
		UpdatedAt:      time.Now().Unix(),
		Status:         ContractDraft,
		Attachments:    attachments,
	}

	contract.ID = contract.HashContract()
	err := contract.ApproveContract(w)
	Handle(err)

	log.Printf("[CONTRACT] Contract created - ID: %x, Title: %s", contract.ID, title)

	return &contract
}

func (contract *Contract) ApproveContract(wallet *wallet.Wallet) error {
	/*
		Approves the contract by verifying the party's signature.
	*/
	if contract.IsCoinbaseOp() {
		return nil
	}

	if contract.Status != ContractDraft {
		log.Printf("[CONTRACT] ✗ Cannot approve contract %x - invalid status: %s", contract.ID, contract.Status)
		return fmt.Errorf("[CONTRACT] Cannot approve contract - invalid status: %s", contract.Status)
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

	err := contract.SignContract(&wallet.PrivateKey, wallet.Address())
	if err != nil {
		log.Printf("[CONTRACT] Failed to approve contract: %x by: %x", contract.ID, wallet.Address())
		return fmt.Errorf("[CONTRACT] Failed to approve contract: %x by: %x", contract.ID, wallet.Address())
	}
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

func (contract *Contract) ApproveMilestone(wallet *wallet.Wallet, milestoneID []byte, evidence []byte) (*Contract, error) {
	/*
		Approves the milestone by verifying the party's signature.
	*/

	if contract.IsCoinbaseOp() {
		return nil, fmt.Errorf("[MILESTONE] ✗ Coinbase contract has no milestones")
	}

	if contract.Status != ContractActive {
		log.Printf("[MILESTONE] ✗ Cannot approve milestone in contract %x - invalid contract status: %s", contract.ID, contract.Status)
		return nil, fmt.Errorf("[MILESTONE] ✗ Cannot approve milestone - invalid contract status: %s", contract.Status)
	}

	milestone, err := contract.GetMilestoneByID(milestoneID)
	if err != nil {
		log.Printf("[MILESTONE] ✗ Milestone %x not found in contract %x", milestoneID, contract.ID)
		return nil, fmt.Errorf("[MILESTONE] ✗ Milestone %x not found in contract %x", milestoneID, contract.ID)
	}

	if milestone.Status == MilestoneCompleted || milestone.Status == MilestoneCancelled {
		log.Printf("[MILESTONE] ✗ Cannot approve milestone %x - invalid status: %s", milestone.ID, milestone.Status)
		return nil, fmt.Errorf("[MILESTONE] Cannot approve milestone - invalid status: %s", milestone.Status)
	}

	if milestone.PartyApproved(string(wallet.Address())) {
		log.Printf("[MILESTONE] ✓ Party %x has already approved the milestone %x", wallet.Address(), milestone.ID)
		return contract, nil
	}

	if evidence != nil {
		milestone.Evidence = evidence
	}

	milestone.ApprovedBy = append(milestone.ApprovedBy, string(wallet.Address()))

	if len(milestone.ApprovedBy) == len(contract.Parties) {
		milestone.Status = MilestoneCompleted
		milestone.CompletedAt = time.Now().Unix()
		contract.UpdatedAt = time.Now().Unix()
		log.Printf("[MILESTONE] ✓ Milestone %x completed - all parties approved", milestone.ID)
	} else {
		pendingApprovals := len(contract.Parties) - len(milestone.ApprovedBy)
		log.Printf("[MILESTONE] ⧗ Milestone %x approved by party %x - %d more approvals needed", milestone.ID, wallet.Address(), pendingApprovals)
	}

	if contract.AllMilestonesCompleted() {
		contract.Status = ContractCompleted
		contract.UpdatedAt = time.Now().Unix()
		log.Printf("[CONTRACT] ✓ Contract %x completed - all milestones completed", contract.ID)
	}

	return contract, nil
}

func (contract *Contract) GetMilestoneByID(milestoneID []byte) (*Milestone, error) {
	/*
		Retrieves a milestone from the contract by its ID.
		Returns the milestone and nil error if found, otherwise returns an error.
	*/
	for _, milestone := range contract.Milestones {
		if bytes.Equal(milestone.ID, milestoneID) {
			return milestone, nil
		}
	}
	return nil, errors.New("Milestone not found")
}

func (milestone *Milestone) PartyApproved(address string) bool {
	/*
		Checks if a specific party has approved the milestone.
	*/
	return slices.Contains(milestone.ApprovedBy, address)
}

func (contract *Contract) AllMilestonesCompleted() bool {
	/*
		Checks if all milestones in the contract are completed.
		Returns true only if every milestone has status MilestoneCompleted.
	*/
	for _, milestone := range contract.Milestones {
		if milestone.Status != MilestoneCompleted {
			return false
		}
	}
	return true
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
