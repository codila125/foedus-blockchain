// Package blockchain implements a simple blockchain with basic functionalities.
package blockchain

import (
	"log"
	"time"

	"github.com/codila125/foedus-blockchain/protobuf"
	"google.golang.org/protobuf/proto"
)

type Block struct {
	Timestamp    int64          // Timestamp of the block creation
	Hash         []byte         // Hash of the block
	Transactions []*Transaction // List of transactions included in the block
	Contracts    []*Contract    // List of contracts included in the block
	PrevHash     []byte         // Hash of the previous block
	Nonce        int            // Nonce used for mining
	Height       int            // Height of the block in the blockchain
}

func CreateBlock(txs []*Transaction, cts []*Contract, prevhash []byte, height int) *Block {
	/*
		Creates a new block with the given contracts, transactions, and previous block hash.
	*/
	block := &Block{
		Timestamp:    time.Now().Unix(),
		Transactions: txs,
		Contracts:    cts,
		PrevHash:     prevhash,
		Nonce:        0,      // Nonce is the number which will be found by Proof of Work
		Height:       height, // Height will be set when adding the block to the blockchain
	}
	pow := NewProof(block)   // Create a new Proof of Work for the block
	nonce, hash := pow.Run() // Run the Proof of Work to find a valid nonce and hash

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func (b *Block) HashTransactions() []byte {
	/*
		Computes the Merkle root of the block's transactions.
		Returns the Merkle root as a byte slice.
	*/
	var transactions [][]byte

	// Serialize each transaction and collect them
	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.ID)
	}
	if len(transactions) == 0 {
		return []byte{}
	}
	tree := NewMerkleTree(transactions) // Create a new Merkle tree from the transactions

	return tree.RootNode.Data
}

func (b *Block) HashContracts() []byte {
	/*
		Computes the Merkle root of the block's contracts.
		Returns the Merkle root as a byte slice.
	*/
	var contracts [][]byte

	// Serialize each contract and collect them
	for _, ct := range b.Contracts {
		contracts = append(contracts, ct.ID)
	}
	if len(contracts) == 0 {
		return []byte{}
	}

	tree := NewMerkleTree(contracts) // Create a new Merkle tree from the contracts

	return tree.RootNode.Data
}

func Genesis(coinbase *Transaction, contractbase *Contract) *Block {
	/*
		Creates the genesis block with a coinbase contract.
		'coinbase' is the coinbase contract that rewards the miner.
	*/
	return CreateBlock([]*Transaction{coinbase}, []*Contract{contractbase}, []byte{}, 0)
}

func (b *Block) SerializeBlock() []byte {
	/*
		Serializes the block into a byte slice.
	*/

	protoBlock := &protobuf.Block{
		Timestamp: b.Timestamp,
		Hash:      b.Hash,
		PrevHash:  b.PrevHash,
		Nonce:     int32(b.Nonce),
		Height:    int32(b.Height),
	}

	for _, tx := range b.Transactions {
		protoTx := &protobuf.Transaction{
			Id:      tx.ID,
			Inputs:  []*protobuf.TxInput{},
			Outputs: []*protobuf.TxOutput{},
		}

		for _, in := range tx.Inputs {
			protoIn := &protobuf.TxInput{
				Id:        in.ID,
				Out:       int32(in.Out),
				Signature: in.Signature,
				PubKey:    in.PubKey,
			}
			protoTx.Inputs = append(protoTx.Inputs, protoIn)
		}

		for _, out := range tx.Outputs {
			protoOut := &protobuf.TxOutput{
				Value:      int32(out.Value),
				PubKeyHash: out.PubKeyHash,
			}
			protoTx.Outputs = append(protoTx.Outputs, protoOut)
		}

		protoBlock.Transactions = append(protoBlock.Transactions, protoTx)
	}

	for _, ct := range b.Contracts {
		protoCt := &protobuf.Contract{
			Id:             ct.ID,
			Title:          ct.Title,
			Description:    ct.Description,
			CreatorAddress: ct.CreatorAddress,
			Attachments:    ct.Attachments,
			Terms:          ct.Terms,
			Status:         string(ct.Status),
			CreatedAt:      ct.CreatedAt,
			UpdatedAt:      ct.UpdatedAt,
			Milestones:     []*protobuf.Milestone{},
			Parties:        []*protobuf.Party{},
		}

		for _, m := range ct.Milestones {
			protoMilestone := &protobuf.Milestone{
				Id:          m.ID,
				Title:       m.Title,
				Description: m.Description,
				Value:       int32(m.Value),
				DueDate:     m.DueDate,
				Status:      string(m.Status),
				CreatedAt:   m.CreatedAt,
				CompletedAt: m.CompletedAt,
				Evidence:    m.Evidence,
				ApprovedBy:  m.ApprovedBy,
			}
			protoCt.Milestones = append(protoCt.Milestones, protoMilestone)
		}

		for _, p := range ct.Parties {
			protoParty := &protobuf.Party{
				Address:   p.Address,
				Role:      string(p.Role),
				PublicKey: p.PublicKey,
				Signature: p.Signature,
			}
			protoCt.Parties = append(protoCt.Parties, protoParty)
		}

		protoBlock.Contracts = append(protoBlock.Contracts, protoCt)
	}

	data, err := proto.Marshal(protoBlock)
	Handle(err)

	return data
}

func DeserializeBlock(data []byte) *Block {
	/*
		Deserializes a byte slice into a Block.
	*/
	var block Block

	protoBlock := &protobuf.Block{}
	err := proto.Unmarshal(data, protoBlock)
	Handle(err)

	block.Timestamp = protoBlock.Timestamp
	block.Hash = protoBlock.Hash
	block.PrevHash = protoBlock.PrevHash
	block.Nonce = int(protoBlock.Nonce)
	block.Height = int(protoBlock.Height)

	for _, protoTx := range protoBlock.Transactions {
		tx := &Transaction{
			ID:      protoTx.Id,
			Inputs:  []TxInput{},
			Outputs: []TxOutput{},
		}

		for _, protoIn := range protoTx.Inputs {
			in := TxInput{
				ID:        protoIn.Id,
				Out:       int(protoIn.Out),
				Signature: protoIn.Signature,
				PubKey:    protoIn.PubKey,
			}
			tx.Inputs = append(tx.Inputs, in)
		}

		for _, protoOut := range protoTx.Outputs {
			out := TxOutput{
				Value:      int(protoOut.Value),
				PubKeyHash: protoOut.PubKeyHash,
			}
			tx.Outputs = append(tx.Outputs, out)
		}

		block.Transactions = append(block.Transactions, tx)
	}

	for _, protoCt := range protoBlock.Contracts {
		ct := &Contract{
			ID:             protoCt.Id,
			Title:          protoCt.Title,
			Description:    protoCt.Description,
			CreatorAddress: protoCt.CreatorAddress,
			Attachments:    protoCt.Attachments,
			Terms:          protoCt.Terms,
			Status:         ContractStatus(protoCt.Status),
			CreatedAt:      protoCt.CreatedAt,
			UpdatedAt:      protoCt.UpdatedAt,
			Milestones:     []*Milestone{},
			Parties:        []*Party{},
		}

		for _, protoM := range protoCt.Milestones {
			m := &Milestone{
				ID:          protoM.Id,
				Title:       protoM.Title,
				Description: protoM.Description,
				Value:       int(protoM.Value),
				DueDate:     protoM.DueDate,
				Status:      MilestoneStatus(protoM.Status),
				CreatedAt:   protoM.CreatedAt,
				CompletedAt: protoM.CompletedAt,
				Evidence:    protoM.Evidence,
				ApprovedBy:  protoM.ApprovedBy,
			}
			ct.Milestones = append(ct.Milestones, m)
		}

		for _, protoP := range protoCt.Parties {
			p := &Party{
				Address:   protoP.Address,
				Role:      ContractRole(protoP.Role),
				PublicKey: protoP.PublicKey,
				Signature: protoP.Signature,
			}
			ct.Parties = append(ct.Parties, p)
		}

		block.Contracts = append(block.Contracts, ct)
	}

	return &block
}

func Handle(err error) {
	/*
		Handles errors by logging them.
	*/
	if err != nil {
		log.Println(err)
	}
}
