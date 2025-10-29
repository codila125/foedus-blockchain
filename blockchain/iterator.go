// Package blockchain provides an iterator for traversing the blockchain from the
// most recent block to the genesis block. This allows for sequential access to
// the blocks in reverse chronological order.
package blockchain

// Iterator returns a new BlockChainIterator instance, initialized to start at
// the latest block in the blockchain. This iterator is the primary mechanism for
// traversing the chain backwards.
func (blockchain *BlockChain) Iterator() *BlockChainIterator {
	iterator := &BlockChainIterator{blockchain.LastHash, blockchain.Database}
	return iterator
}

// Next fetches the block currently pointed to by the iterator, deserializes it,
// and then updates the iterator to point to the previous block in the chain.
// This method allows for iterating through the entire blockchain, one block at a
// time, from newest to oldest. It returns a pointer to the deserialized block.
func (iter *BlockChainIterator) Next() *Block {
	db := iter.Database.GetRawDB()

	blockData, closer, err := db.Get(iter.CurrentHash)
	if err != nil {
		Handle(err)
	}

	blockDataCopy := make([]byte, len(blockData))
	copy(blockDataCopy, blockData)
	closer.Close()

	block := DeserializeBlock(blockDataCopy)

	iter.CurrentHash = block.PrevHash

	return block
}
