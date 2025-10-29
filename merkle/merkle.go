// Package merkle implements a Merkle tree data structure for efficient data integrity
// verification and cryptographic proofs in the blockchain.
package merkle

import (
	"crypto/sha256"
)

// MerkleTree represents a binary hash tree where each leaf node contains data
// and each non-leaf node contains the hash of its children.
type MerkleTree struct {
	RootNode *MerkleNode
}

// MerkleNode represents a single node in the Merkle tree, containing either
// data (for leaf nodes) or the combined hash of its children (for internal nodes).
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

// NewMerkleNode creates a new Merkle tree node. For leaf nodes (no children),
// it hashes the provided data. For internal nodes, it hashes the concatenation
// of the children's hashes.
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	node := MerkleNode{}

	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		node.Data = hash[:]
	} else {
		prevHashes := append(left.Data, right.Data...)
		hash := sha256.Sum256(prevHashes)
		node.Data = hash[:]
	}
	node.Left = left
	node.Right = right

	return &node
}

// NewMerkleTree constructs a Merkle tree from an array of data items.
// If the number of data items is odd, the last item is duplicated to ensure
// all levels have an even number of nodes for proper pairing.
func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode

	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	for _, dat := range data {
		node := NewMerkleNode(nil, nil, dat)
		nodes = append(nodes, *node)
	}

	for i := 0; i < len(data)/2; i++ {
		var level []MerkleNode

		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			level = append(level, *node)
		}

		nodes = level
	}
	return &MerkleTree{&nodes[0]}
}
