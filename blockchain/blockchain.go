package blockchain

import (
	"encoding/hex"
	"fmt"
	"os"
	"runtime"

	badger "github.com/dgraph-io/badger/v4"
)

const (
	dbPath      = "./tmp/blocks"
	dbFile      = "./tmp/blocks/MANIFEST"
	genesisData = "First Transaction from Genesis"
)

type BlockChain struct {
	LastHash []byte
	Database *badger.DB
}

type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func DBExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func NewBlockChain() *BlockChain {
	var lastHash []byte

	if DBExists() {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}

	options := badger.DefaultOptions(dbPath)

	db, err := badger.Open(options)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		cbtx := CoinbaseTx("satoshi", genesisData)
		genesis := Genesis(cbtx)
		fmt.Println("Genesis created")
		err := txn.Set(genesis.Hash, genesis.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), genesis.Hash)

		lastHash = genesis.Hash

		return err
	})

	Handle(err)

	blockchain := BlockChain{lastHash, db}
	return &blockchain
}

func ContinueBlockChain() *BlockChain {
	if !DBExists() {
		fmt.Println("Blockchain does not exist")
		runtime.Goexit()
	}

	var lastHash []byte

	options := badger.DefaultOptions(dbPath)
	db, err := badger.Open(options)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.ValueCopy(nil)

		return err
	})

	Handle(err)

	chain := BlockChain{lastHash, db}
	return &chain
}

func (blockchain *BlockChain) AddBlock(transactions []*Transcation) {
	var lastHash []byte

	err := blockchain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.ValueCopy(nil)

		return err
	})

	Handle(err)

	newBlock := CreateBlock(transactions, lastHash)
	err = blockchain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)

		blockchain.LastHash = newBlock.Hash
		return err
	})

	Handle(err)
}

func (blockchain *BlockChain) Iterator() *BlockChainIterator {
	iterator := &BlockChainIterator{blockchain.LastHash, blockchain.Database}
	return iterator
}

func (iter *BlockChainIterator) Next() *Block {
	var block *Block

	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		Handle(err)
		encodedBlock, err := item.ValueCopy(nil)
		block = Deserialize(encodedBlock)

		return err
	})

	Handle(err)

	iter.CurrentHash = block.PrevHash

	return block
}

func (blockchain *BlockChain) FindUnspentTransactions(address string) []Transcation {
	var unspentTXs []Transcation
	spentTXs := make(map[string][]int)

	iterator := blockchain.Iterator()

	for {
		block := iterator.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Output {
				if spentTXs[txID] != nil {
					for _, spentOut := range spentTXs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}

			}
			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.ID)
						spentTXs[inTxID] = append(spentTXs[inTxID], in.Out)
					}
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}

	}
	return unspentTXs
}

func (blockchain *BlockChain) FindUTXO(address string) []TxOutput {
	var UTXOs []TxOutput
	unspentTransactions := blockchain.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Output {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

func (blockchain *BlockChain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := blockchain.FindUnspentTransactions(address)
	acumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Output {
			if out.CanBeUnlockedWith(address) && acumulated < amount {
				acumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if acumulated >= amount {
					break Work
				}
			}
		}
	}

	return acumulated, unspentOutputs
}
