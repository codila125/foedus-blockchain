package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"log"
	"math"
	"math/big"
)

const Difficulty = 2 // Difficulty is the number of leading zero bits required in the hash

type ProofOfWork struct {
	Block  *Block   // The block to be mined
	Target *big.Int // The target value that the hash must be less than
}

func NewProof(b *Block) *ProofOfWork {
	/*
		Creates a new Proof of Work instance for the given block.
	*/
	target := big.NewInt(1)                  // Initialize target to 1
	target.Lsh(target, uint(256-Difficulty)) // Left shift the target to set the difficulty

	pow := &ProofOfWork{b, target}

	return pow
}

func (pow *ProofOfWork) InitData(nonce int) []byte {
	/*
		Initializes the data for the Proof of Work algorithm with the given nonce.
	*/
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevHash,
			pow.Block.HashTransactions(), // Merkle root of transactions
			pow.Block.HashContracts(),    // Merkle root of contracts
			ToHex(int64(nonce)),
			ToHex(int64(Difficulty)),
		},
		[]byte{},
	)
	return data
}

func (pow *ProofOfWork) Validate() bool {
	/*
		Validates the Proof of Work by checking if the hash of the block is less than the target.
		Returns true if valid, false otherwise.
	*/
	var intHash big.Int

	data := pow.InitData(pow.Block.Nonce)

	hash := sha256.Sum256(data)
	intHash.SetBytes(hash[:])

	return intHash.Cmp(pow.Target) == -1
}

func ToHex(num int64) []byte {
	/*
		Converts an int64 number to a byte slice in big-endian order.
	*/
	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.BigEndian, int64(num))
	if err != nil {
		log.Panic(err)
	}

	return buffer.Bytes()
}

func (pow *ProofOfWork) Run() (int, []byte) {
    var intHash big.Int
    var hash [32]byte

    nonce := 0
    progressInterval := 100000  // Log every 100k attempts

    // Iterate until a valid nonce is found or the maximum integer value is reached
    for nonce < math.MaxInt64 {
        data := pow.InitData(nonce)
        hash = sha256.Sum256(data)

        intHash.SetBytes(hash[:])

        // Show progress every progressInterval attempts
        if nonce%progressInterval == 0 && nonce > 0 {
            log.Printf("[PoW] Mining in progress... Nonce: %d (attempts)", nonce)
        }

        // Check if the hash meets the target
        if intHash.Cmp(pow.Target) == -1 {
            log.Printf("[PoW] ✓ Mining complete! Found valid nonce: %d", nonce)
            break
        } else {
            nonce++
        }
    }

    return nonce, hash[:]
}
