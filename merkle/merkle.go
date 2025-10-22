// Package merkle implements a simple Merkle tree structure for data integrity verification.
package merkle

import (
	"crypto/sha256"
)

type MerkleTree struct {
	RootNode *MerkleNode // The root node of the Merkle tree
}

type MerkleNode struct {
	Left  *MerkleNode // Left child node
	Right *MerkleNode // Right child node
	Data  []byte      // Hash data
}

func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	/*
		Creates a new Merkle node.
		'left' and 'right' are the child nodes.
		'data' is the data for leaf nodes; for non-leaf nodes, it is nil.
	*/
	node := MerkleNode{}

	// If there are no child nodes, it's a leaf node
	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		node.Data = hash[:]
	} else {
		// For non-leaf nodes, concatenate the hashes of the child nodes and hash them
		prevHashes := append(left.Data, right.Data...)
		hash := sha256.Sum256(prevHashes)
		node.Data = hash[:]
	}
	node.Left = left
	node.Right = right

	return &node
}

func NewMerkleTree(data [][]byte) *MerkleTree {
	/*
		Creates a new Merkle tree from the given data.
	*/
	var nodes []MerkleNode

	// If the number of data items is odd, duplicate the last item
	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	// Create leaf nodes
	for _, dat := range data {
		node := NewMerkleNode(nil, nil, dat)
		nodes = append(nodes, *node)
	}

	// Build the tree by pairing nodes and creating parent nodes
	for i := 0; i < len(data)/2; i++ {
		var level []MerkleNode

		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			level = append(level, *node) // Append the new parent node to the current level
		}

		nodes = level
	}
	return &MerkleTree{&nodes[0]}
}
