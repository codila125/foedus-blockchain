// Package blockchain implements the Proof-of-Work (PoW) consensus algorithm,
// which is essential for securing the blockchain. PoW requires miners to solve a
// computationally intensive puzzle to add new blocks, thus preventing malicious
// actors from easily altering the chain.
package blockchain

import (
	"crypto/sha256"
	"encoding/binary"
	"log"
	"math"
	"math/big"
)

// difficulty is a constant that determines the complexity of the mining puzzle.
// A higher difficulty requires more computational effort to find a valid hash,
// making the blockchain more secure. This value is used to calculate the target.
const difficulty = 12

// ProofOfWork encapsulates the data and logic required for the PoW algorithm.
// It holds a reference to the block being mined and the target value that the
// block's hash must be less than.
type ProofOfWork struct {
	Block  *Block
	Target *big.Int
}

// NewProof creates and initializes a new ProofOfWork instance for a given block.
// It calculates the target value based on the predefined difficulty, setting the
// challenge for the miners.
func NewProof(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-difficulty))

	pow := &ProofOfWork{b, target}

	return pow
}

// InitData prepares the data that will be hashed in the PoW process. It combines
// the block's essential headers—such as the previous block's hash, the Merkle roots
// of transactions and contracts, the nonce, and the difficulty—into a single byte slice.
func (pow *ProofOfWork) InitData(nonce int) []byte {
	buffer := make([]byte, 0, 128)
	buffer = append(buffer, pow.Block.PrevHash...)
	buffer = append(buffer, pow.Block.HashTransactions()...)
	buffer = append(buffer, pow.Block.HashContracts()...)

	nonceBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(nonceBytes, uint64(nonce))
	buffer = append(buffer, nonceBytes...)

	difficultyBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(difficultyBytes, uint64(difficulty))
	buffer = append(buffer, difficultyBytes...)

	return buffer
}

// Validate checks if a block's hash meets the PoW requirement. It re-hashes the
// block's data with its stored nonce and compares the result to the target.
// It returns true if the hash is valid, confirming that the required work was done.
func (pow *ProofOfWork) Validate() bool {
	var intHash big.Int

	data := pow.InitData(pow.Block.Nonce)

	hash := sha256.Sum256(data)
	intHash.SetBytes(hash[:])

	return intHash.Cmp(pow.Target) == -1
}

// Run executes the mining process. It repeatedly hashes the block's data with
// different nonce values until it finds a hash that is less than the target.
// This iterative process is the "work" in Proof-of-Work. It returns the successful
// nonce and the resulting block hash.
func (pow *ProofOfWork) Run() (int, []byte) {
	var intHash big.Int
	var hash [32]byte

	nonce := 0
	progressInterval := 100000

	for nonce < math.MaxInt64 {
		data := pow.InitData(nonce)
		hash = sha256.Sum256(data)

		intHash.SetBytes(hash[:])

		if nonce%progressInterval == 0 && nonce > 0 {
			log.Printf("[PoW] Mining in progress... Nonce: %d (attempts)", nonce)
		}

		if intHash.Cmp(pow.Target) == -1 {
			log.Printf("[PoW] ✓ Mining complete! Found valid nonce: %d", nonce)
			break
		} else {
			nonce++
		}
	}

	return nonce, hash[:]
}
