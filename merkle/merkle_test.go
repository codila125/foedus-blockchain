package merkle

import (
	"bytes"
	"crypto/sha256"
	"testing"
)

func TestNewMerkleNode_LeafNode(t *testing.T) {
	data := []byte("test data")
	node := NewMerkleNode(nil, nil, data)

	// Verify the node has no children
	if node.Left != nil {
		t.Error("Leaf node should not have a left child")
	}
	if node.Right != nil {
		t.Error("Leaf node should not have a right child")
	}

	// Verify the hash is computed correctly
	expectedHash := sha256.Sum256(data)
	if !bytes.Equal(node.Data, expectedHash[:]) {
		t.Error("Leaf node hash does not match expected SHA256 hash")
	}
}

func TestNewMerkleNode_InternalNode(t *testing.T) {
	leftData := []byte("left data")
	rightData := []byte("right data")

	leftNode := NewMerkleNode(nil, nil, leftData)
	rightNode := NewMerkleNode(nil, nil, rightData)
	parentNode := NewMerkleNode(leftNode, rightNode, nil)

	// Verify the parent has correct children
	if parentNode.Left != leftNode {
		t.Error("Parent node should have correct left child")
	}
	if parentNode.Right != rightNode {
		t.Error("Parent node should have correct right child")
	}

	// Verify the parent hash is computed from children's hashes
	combinedHashes := append(leftNode.Data, rightNode.Data...)
	expectedHash := sha256.Sum256(combinedHashes)
	if !bytes.Equal(parentNode.Data, expectedHash[:]) {
		t.Error("Internal node hash does not match expected combined hash")
	}
}

func TestNewMerkleTree_SingleElement(t *testing.T) {
	data := [][]byte{
		[]byte("single element"),
	}

	tree := NewMerkleTree(data)

	if tree.RootNode == nil {
		t.Fatal("Tree root should not be nil")
	}

	// With a single element, it gets duplicated, so root should have children
	if tree.RootNode.Left == nil || tree.RootNode.Right == nil {
		t.Error("Single element tree should have duplicated leaf nodes")
	}

	// Both children should have the same hash (since data was duplicated)
	if !bytes.Equal(tree.RootNode.Left.Data, tree.RootNode.Right.Data) {
		t.Error("Duplicated leaf nodes should have identical hashes")
	}
}

func TestNewMerkleTree_TwoElements(t *testing.T) {
	data := [][]byte{
		[]byte("element1"),
		[]byte("element2"),
	}

	tree := NewMerkleTree(data)

	if tree.RootNode == nil {
		t.Fatal("Tree root should not be nil")
	}

	// Verify root has two children
	if tree.RootNode.Left == nil {
		t.Error("Root should have a left child")
	}
	if tree.RootNode.Right == nil {
		t.Error("Root should have a right child")
	}

	// Verify leaf hashes
	expectedLeftHash := sha256.Sum256(data[0])
	if !bytes.Equal(tree.RootNode.Left.Data, expectedLeftHash[:]) {
		t.Error("Left leaf hash is incorrect")
	}

	expectedRightHash := sha256.Sum256(data[1])
	if !bytes.Equal(tree.RootNode.Right.Data, expectedRightHash[:]) {
		t.Error("Right leaf hash is incorrect")
	}

	// Verify root hash
	combinedHashes := append(expectedLeftHash[:], expectedRightHash[:]...)
	expectedRootHash := sha256.Sum256(combinedHashes)
	if !bytes.Equal(tree.RootNode.Data, expectedRootHash[:]) {
		t.Error("Root hash is incorrect")
	}
}

func TestNewMerkleTree_FourElements(t *testing.T) {
	data := [][]byte{
		[]byte("a"),
		[]byte("b"),
		[]byte("c"),
		[]byte("d"),
	}

	tree := NewMerkleTree(data)

	if tree.RootNode == nil {
		t.Fatal("Tree root should not be nil")
	}

	// With 4 elements, we should have:
	// Level 0: 4 leaf nodes
	// Level 1: 2 internal nodes
	// Level 2: 1 root node

	// Calculate expected hashes manually
	hashA := sha256.Sum256(data[0])
	hashB := sha256.Sum256(data[1])
	hashC := sha256.Sum256(data[2])
	hashD := sha256.Sum256(data[3])

	hashAB := sha256.Sum256(append(hashA[:], hashB[:]...))
	hashCD := sha256.Sum256(append(hashC[:], hashD[:]...))

	expectedRoot := sha256.Sum256(append(hashAB[:], hashCD[:]...))

	if !bytes.Equal(tree.RootNode.Data, expectedRoot[:]) {
		t.Error("Root hash for 4-element tree is incorrect")
	}
}

func TestNewMerkleTree_OddElements(t *testing.T) {
	data := [][]byte{
		[]byte("a"),
		[]byte("b"),
		[]byte("c"),
	}

	tree := NewMerkleTree(data)

	if tree.RootNode == nil {
		t.Fatal("Tree root should not be nil")
	}

	// With 3 elements, the last one is duplicated to make 4
	// So we expect: a, b, c, c
	hashA := sha256.Sum256(data[0])
	hashB := sha256.Sum256(data[1])
	hashC := sha256.Sum256(data[2])

	hashAB := sha256.Sum256(append(hashA[:], hashB[:]...))
	hashCC := sha256.Sum256(append(hashC[:], hashC[:]...))

	expectedRoot := sha256.Sum256(append(hashAB[:], hashCC[:]...))

	if !bytes.Equal(tree.RootNode.Data, expectedRoot[:]) {
		t.Error("Root hash for odd-element tree is incorrect")
	}
}

func TestNewMerkleTree_Deterministic(t *testing.T) {
	data := [][]byte{
		[]byte("test1"),
		[]byte("test2"),
		[]byte("test3"),
		[]byte("test4"),
	}

	tree1 := NewMerkleTree(data)
	tree2 := NewMerkleTree(data)

	if !bytes.Equal(tree1.RootNode.Data, tree2.RootNode.Data) {
		t.Error("Trees with same data should produce identical root hashes")
	}
}

func TestNewMerkleTree_DifferentDataDifferentRoot(t *testing.T) {
	data1 := [][]byte{
		[]byte("a"),
		[]byte("b"),
	}

	data2 := [][]byte{
		[]byte("c"),
		[]byte("d"),
	}

	tree1 := NewMerkleTree(data1)
	tree2 := NewMerkleTree(data2)

	if bytes.Equal(tree1.RootNode.Data, tree2.RootNode.Data) {
		t.Error("Trees with different data should produce different root hashes")
	}
}

func TestNewMerkleTree_OrderMatters(t *testing.T) {
	data1 := [][]byte{
		[]byte("a"),
		[]byte("b"),
	}

	data2 := [][]byte{
		[]byte("b"),
		[]byte("a"),
	}

	tree1 := NewMerkleTree(data1)
	tree2 := NewMerkleTree(data2)

	if bytes.Equal(tree1.RootNode.Data, tree2.RootNode.Data) {
		t.Error("Trees with reordered data should produce different root hashes")
	}
}

func TestMerkleNode_HashLength(t *testing.T) {
	data := []byte("test")
	node := NewMerkleNode(nil, nil, data)

	// SHA256 produces 32-byte hashes
	if len(node.Data) != 32 {
		t.Errorf("Expected hash length of 32, got %d", len(node.Data))
	}
}

func TestNewMerkleTree_LargeDataset(t *testing.T) {
	// Test with a larger dataset (power of 2)
	data := make([][]byte, 16)
	for i := range 16 {
		data[i] = []byte{byte(i)}
	}

	tree := NewMerkleTree(data)

	if tree.RootNode == nil {
		t.Fatal("Tree root should not be nil for large dataset")
	}

	// Verify root hash is 32 bytes
	if len(tree.RootNode.Data) != 32 {
		t.Errorf("Root hash should be 32 bytes, got %d", len(tree.RootNode.Data))
	}
}

func TestNewMerkleTree_LargeOddDataset(t *testing.T) {
	// Test with a larger odd dataset
	data := make([][]byte, 17)
	for i := range 17 {
		data[i] = []byte{byte(i)}
	}

	tree := NewMerkleTree(data)

	if tree.RootNode == nil {
		t.Fatal("Tree root should not be nil for large odd dataset")
	}

	// Verify root hash is 32 bytes
	if len(tree.RootNode.Data) != 32 {
		t.Errorf("Root hash should be 32 bytes, got %d", len(tree.RootNode.Data))
	}
}

func TestNewMerkleTree_EmptyData(t *testing.T) {
	// Test with empty byte slices as data
	data := [][]byte{
		{},
		{},
	}

	tree := NewMerkleTree(data)

	if tree.RootNode == nil {
		t.Fatal("Tree root should not be nil even with empty data elements")
	}

	// Both leaf hashes should be identical (hash of empty data)
	if !bytes.Equal(tree.RootNode.Left.Data, tree.RootNode.Right.Data) {
		t.Error("Leaf nodes with empty data should have identical hashes")
	}
}

// Benchmark tests
func BenchmarkNewMerkleNode_Leaf(b *testing.B) {
	data := []byte("benchmark data")

	for b.Loop() {
		NewMerkleNode(nil, nil, data)
	}
}

func BenchmarkNewMerkleTree_Small(b *testing.B) {
	data := [][]byte{
		[]byte("a"),
		[]byte("b"),
		[]byte("c"),
		[]byte("d"),
	}

	for b.Loop() {
		NewMerkleTree(data)
	}
}

func BenchmarkNewMerkleTree_Medium(b *testing.B) {
	data := make([][]byte, 64)
	for i := range 64 {
		data[i] = []byte{byte(i)}
	}

	for b.Loop() {
		NewMerkleTree(data)
	}
}

func BenchmarkNewMerkleTree_Large(b *testing.B) {
	data := make([][]byte, 1024)
	for i := range 1024 {
		data[i] = []byte{byte(i % 256)}
	}

	for b.Loop() {
		NewMerkleTree(data)
	}
}
