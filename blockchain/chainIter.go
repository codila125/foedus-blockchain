package blockchain


func (blockchain *BlockChain) Iterator() *BlockChainIterator {
	/*
		Creates and returns a new BlockChainIterator starting from the last block in the chain.
	*/
	iterator := &BlockChainIterator{blockchain.LastHash, blockchain.Database}
	return iterator
}

func (iter *BlockChainIterator) Next() *Block {
	/*
		Fetch the next block in the chain using the current hash stored in the iterator.
		Update the iterator's current hash to the previous block's hash for the next call.
		Return the deserialized block.
	*/
    db := iter.Database.GetRawDB()
    
    blockData, closer, err := db.Get(iter.CurrentHash)
    if err != nil {
        Handle(err)
    }
    defer closer.Close()

    // Copy data before closing the closer
    blockDataCopy := make([]byte, len(blockData))
    copy(blockDataCopy, blockData)

    block := Deserialize(blockDataCopy)

    // Move to the previous block for next iteration
    iter.CurrentHash = block.PrevHash

    return block
}
