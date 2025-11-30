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

// CoinbaseOp creates a special "coinbase" contract, which represents the mining reward
// for creating a new block. This contract has no parties or milestones and serves as the
// foundational transaction in a genesis block or as a reward in subsequent blocks.
// The 'data' field can be arbitrary but is typically used for miner-specific information.
func CoinbaseOp(creator, data string) *Contract {
	if data == "" {
		randData := make([]byte, 20)
		_, err := rand.Read(randData)
		if err != nil {
			log.Print(err)
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

// IsCoinbaseOp checks if a contract is a coinbase operation. A coinbase contract is
// identified by its specific title and the absence of milestones and parties.
func (contract *Contract) IsCoinbaseOp() bool {
	return contract.Title == "Coinbase Foedus" && len(contract.Milestones) == 0 && len(contract.Parties) == 0
}

// FindContract searches the entire blockchain for a contract by its ID. It optimizes the
// search by first checking the Incomplete Contract (ICCT) set for active contracts.
// If not found, it performs a full scan of the blockchain history.
func (blockchain *BlockChain) FindContract(contractID string) (Contract, error) {
	targetID, err := hex.DecodeString(contractID)
	if err != nil {
		targetID = []byte(contractID)
	}

	icctSet := ICCTSet{blockchain}
	contract, err := icctSet.GetContract(targetID)
	if err == nil && contract != nil {
		log.Printf("[CONTRACT] Contract %x found in ICCT set - Status: %s", targetID, contract.Status)
		return *contract, nil
	}

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

// SignContract generates a cryptographic signature for the contract using a party's private key.
// This signature serves as proof of agreement from that party. The contract's hash is signed,
// and the resulting signature is stored in the corresponding party's entry within the contract.
func (contract *Contract) SignContract(privKey *ed25519.PrivateKey, partyAddress []byte) error {
	if contract.IsCoinbaseOp() {
		return nil
	}

	ctCopy := *contract
	ctCopy.Parties = make([]*Party, len(contract.Parties))
	for i, party := range contract.Parties {
		partyCopy := *party
		partyCopy.Signature = nil
		ctCopy.Parties[i] = &partyCopy
	}

	contractHash := ctCopy.HashContract()

	signature := ed25519.Sign(*privKey, contractHash)

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

// VerifyContract checks the validity of a contract based on its current status.
// - For DRAFT contracts, no signatures are required.
// - For ACTIVE contracts, it ensures that all participating parties have provided a valid signature.
// This function is crucial for ensuring that only valid contracts are included in new blocks.
func (blockchain *BlockChain) VerifyContract(contract *Contract) bool {
	if contract.IsCoinbaseOp() {
		return true
	}

	// DRAFT contracts don't need signatures to be added to blockchain
	if contract.Status == ContractDraft {
		log.Printf("[CONTRACT] ✓ Contract %x verified - Status: DRAFT (awaiting signatures)", contract.ID)
		return true
	}

	if contract.Status == ContractActive {
		if !contract.AreAllPartiesSigned() {
			unsigned := contract.GetUnsignedParties()
			log.Printf("[CONTRACT] ✗ Contract %x cannot be ACTIVE - %d parties still need to sign", contract.ID, len(unsigned))
			return false
		}

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

// VerifyContractSignature validates a single party's signature on the contract. It reconstructs
// the signed data hash and uses the party's public key to verify the signature, ensuring
// that the party indeed authorized the contract.
func (contract *Contract) VerifyContractSignature(partyAddress []byte, pubKeyBytes []byte) bool {
	if contract.IsCoinbaseOp() {
		return true
	}

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

	ctCopy := *contract
	ctCopy.Parties = make([]*Party, len(contract.Parties))
	for i, party := range contract.Parties {
		partyCopy := *party
		partyCopy.Signature = nil
		ctCopy.Parties[i] = &partyCopy
	}

	contractHash := ctCopy.HashContract()

	// Decode signature (r||s format, 64 bytes)
	if len(partySignature) != 64 {
		log.Printf("[CONTRACT] ✗ Invalid signature length: %d (expected 64)", len(partySignature))
		return false
	}

	pubKey := ed25519.PublicKey(pubKeyBytes)

	return ed25519.Verify(pubKey, contractHash, partySignature)
}

// AreAllPartiesSigned iterates through the contract's parties to check if each one has
// provided a signature. It returns true only if every party's signature field is populated.
func (contract *Contract) AreAllPartiesSigned() bool {
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

// GetUnsignedParties returns a slice containing all parties who have not yet signed the contract.
// This is useful for identifying which participants are still required to approve the contract.
func (contract *Contract) GetUnsignedParties() []*Party {
	unsigned := make([]*Party, 0)

	for _, party := range contract.Parties {
		if len(party.Signature) == 0 {
			unsigned = append(unsigned, party)
		}
	}

	return unsigned
}

// CreateContract initializes a new contract with the specified details, including title,
// description, milestones, parties, and terms. It sets the initial status to 'ContractDraft',
// automatically adds the creator to the list of parties, and signs the contract on behalf of the creator.
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

// ApproveContract allows a party to sign and approve a contract. This action is only valid
// when the contract is in the 'ContractDraft' status. If this approval is the final one
// required, the contract's status is automatically transitioned to 'ContractActive'.
func (contract *Contract) ApproveContract(wallet *wallet.Wallet) error {
	if contract.IsCoinbaseOp() {
		return nil
	}

	if contract.Status != ContractDraft {
		log.Printf("[CONTRACT] ✗ Cannot approve contract %x - invalid status: %s", contract.ID, contract.Status)
		return fmt.Errorf("[CONTRACT] Cannot approve contract - invalid status: %s", contract.Status)
	}

	willBeComplete := true
	for _, party := range contract.Parties {
		if bytes.Equal([]byte(party.Address), wallet.Address()) {
			continue
		}
		if len(party.Signature) == 0 {
			willBeComplete = false
			break
		}
	}

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

// ApproveMilestone allows a party to approve a specific milestone within a contract.
// This action is only valid for 'ContractActive' contracts. If all parties approve a milestone,
// its status changes to 'MilestoneCompleted'. If all milestones become completed, the
// entire contract transitions to 'ContractCompleted'.
func (contract *Contract) ApproveMilestone(wallet *wallet.Wallet, milestoneID []byte, evidence []byte) (*Contract, error) {
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

// GetMilestoneByID searches for and returns a milestone within the contract by its unique ID.
// It returns an error if no matching milestone is found.
func (contract *Contract) GetMilestoneByID(milestoneID []byte) (*Milestone, error) {
	for _, milestone := range contract.Milestones {
		if bytes.Equal(milestone.ID, milestoneID) {
			return milestone, nil
		}
	}
	return nil, errors.New("Milestone not found")
}

// PartyApproved checks if a specific party (by address) has already approved the milestone.
func (milestone *Milestone) PartyApproved(address string) bool {
	return slices.Contains(milestone.ApprovedBy, address)
}

// AllMilestonesCompleted checks if all milestones within the contract have reached the
// 'MilestoneCompleted' status. This is a key condition for the contract itself to be
// considered completed.
func (contract *Contract) AllMilestonesCompleted() bool {
	for _, milestone := range contract.Milestones {
		if milestone.Status != MilestoneCompleted {
			return false
		}
	}
	return true
}

// CancelContract allows the creator of the contract to cancel it. This action is only
// valid when the contract is in the 'ContractCompleted' status. Cancelling a contract
// changes its status to 'ContractCancelled' and prevents any further actions on it.
func (contract *Contract) CancelContract(cancellerAddress string) error {
	if contract.IsCoinbaseOp() {
		return fmt.Errorf("[CONTRACT] ✗ Coinbase contract cannot be cancelled")
	}

	if contract.Status == ContractCompleted {
		log.Printf("[CONTRACT] ✗ Cannot cancel contract %x - invalid status: %s", contract.ID, contract.Status)
		return fmt.Errorf("[CONTRACT] ✗ Cannot cancel contract - invalid status: %s", contract.Status)
	}

	if contract.CreatorAddress != cancellerAddress {
		log.Printf("[CONTRACT] ✗ Only the creator can cancel the contract %x", contract.ID)
		return fmt.Errorf("[CONTRACT] ✗ Only the creator can cancel the contract")
	}

	contract.Status = ContractCancelled
	contract.UpdatedAt = time.Now().Unix()
	log.Printf("[CONTRACT] ✓ Contract %x has been cancelled by creator %s", contract.ID, contract.CreatorAddress)

	return nil
}