package blockchain

import (
	badger "github.com/dgraph-io/badger/v4"
)

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
	var block *Block

	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash) // Get the block data for the current hash
		Handle(err)
		encodedBlock, err := item.ValueCopy(nil) // Copy the block data
		block = Deserialize(encodedBlock)        // Deserialize the block

		return err
	})

	Handle(err)

	iter.CurrentHash = block.PrevHash // Move to the previous block for the next iteration

	return block
}
