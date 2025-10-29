// Package blockchain implements the logic for creating, signing, and verifying
// transactions. Transactions are the fundamental building blocks for transferring
// value and executing operations on the Foedus blockchain.
package blockchain

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/codila125/foedus-blockchain/wallet"
)

// Transaction represents a transfer of value on the blockchain. It consists of
// a set of inputs, which reference previously unspent transaction outputs (UTXOs),
// and a set of outputs, which create new UTXOs.
type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

// CoinbaseTx creates a special type of transaction known as a coinbase transaction.
// This transaction is created by a miner who successfully mines a new block and
// serves as a reward. It has no inputs and creates new coins.
func CoinbaseTx(to, data string) *Transaction {
	if data == "" {
		randData := make([]byte, 20)
		_, err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}
		data = fmt.Sprintf("%x", randData)
	}

	txin := TxInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTxOutput(20, to)

	tx := Transaction{nil, []TxInput{txin}, []TxOutput{*txout}}
	tx.ID = tx.HashTransaction()

	return &tx
}

// IsCoinbaseTx checks if a transaction is a coinbase transaction. A coinbase
// transaction is identified by having exactly one input where the referenced
// transaction ID is empty and the output index is -1.
func (tx *Transaction) IsCoinbaseTx() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

// FindTransaction searches the entire blockchain for a transaction with a given ID.
// It iterates through each block and its transactions until a match is found.
// If the transaction is not found, it returns an error.
func (blockchain *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	iterator := blockchain.Iterator()

	for {
		block := iterator.Next()

		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, ID) {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("transaction does not exist")
}

// SignTransaction signs a transaction using the provided private key. Before signing,
// it retrieves all the previous transactions referenced by the inputs to ensure
// the integrity of the signing process. Coinbase transactions are not signed.
func (blockchain *BlockChain) SignTransaction(tx *Transaction, privKey ed25519.PrivateKey) {
	if tx.IsCoinbaseTx() {
		return
	}
	prevTXs := make(map[string]Transaction)
	for _, in := range tx.Inputs {
		prevTX, err := blockchain.FindTransaction(in.ID)
		Handle(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	tx.Sign(privKey, prevTXs)
}

// VerifyTransaction validates the signatures of a transaction. It checks that each
// input's signature is valid for the public key of the referenced output. It can
// also verify transactions that are part of the same block or mempool using txMap.
func (blockchain *BlockChain) VerifyTransaction(tx *Transaction, txMap map[string]Transaction) bool {
	if tx.IsCoinbaseTx() {
		return true
	}
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		if prevTx, ok := txMap[hex.EncodeToString(in.ID)]; ok {
			prevTXs[hex.EncodeToString(prevTx.ID)] = prevTx
		} else {
			prevTX, err := blockchain.FindTransaction(in.ID)
			if err != nil {
				log.Printf("[VERIFY] Transaction verification failed - Parent transaction %x not found", in.ID)
				return false
			}
			prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
		}
	}

	if !tx.Verify(prevTXs) {
		log.Printf("[VERIFY] Transaction %x signature verification failed", tx.ID)
		return false
	}

	return true
}

// NewTransaction creates a new transaction to transfer a specified amount from a
// sender's wallet to a recipient's address. It gathers spendable outputs from the
// UTXO set, creates the necessary inputs and outputs, and signs the transaction.
// It panics if the sender has insufficient funds.
func NewTransaction(w *wallet.Wallet, to string, amount int, UTXO *UTXOSet) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	pubKeyHash := wallet.PublicKeyHash(w.PublicKey)
	acc, validOutputs := UTXO.FindSpendableOutputs(pubKeyHash, amount)

	if acc < amount {
		log.Panicf("[TRANSACTION] Insufficient funds - Required: %d, Available: %d", amount, acc)
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		Handle(err)

		for _, out := range outs {
			input := TxInput{txID, out, nil, w.PublicKey}
			inputs = append(inputs, input)
		}
	}

	from := string(w.Address())
	outputs = append(outputs, *NewTxOutput(amount, to))

	if acc > amount {
		outputs = append(outputs, *NewTxOutput(acc-amount, from))
	}

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.HashTransaction()

	UTXO.Blockchain.SignTransaction(&tx, w.PrivateKey)

	log.Printf("[TRANSACTION] Transaction created - ID: %x, Amount: %d", tx.ID, amount)
	return &tx
}

// Sign generates a digital signature for each input in the transaction. It uses
// the provided private key and a map of the previous transactions to create a
// trimmed copy of the transaction for signing, ensuring each input is authorized.
func (tx *Transaction) Sign(privKey ed25519.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbaseTx() {
		return
	}

	for _, in := range tx.Inputs {
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panicf("[TRANSACTION] Signing failed - Invalid parent transaction: %x", in.ID)
		}
	}

	txCopy := tx.TrimmedCopy()

	for inID, in := range txCopy.Inputs {
		prevTX := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inID].Signature = nil
		txCopy.Inputs[inID].PubKey = prevTX.Outputs[in.Out].PubKeyHash

		dataToSign := fmt.Sprintf("%x\n", txCopy)

		signature := ed25519.Sign(privKey, []byte(dataToSign))
		tx.Inputs[inID].Signature = signature
		txCopy.Inputs[inID].PubKey = nil
	}
}

// TrimmedCopy creates a simplified copy of the transaction, with signatures and
// public keys removed from the inputs. This version of the transaction is used
// during the signing and verification process to ensure consistency.
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	for _, in := range tx.Inputs {
		inputs = append(inputs, TxInput{in.ID, in.Out, nil, nil})
	}

	for _, out := range tx.Outputs {
		outputs = append(outputs, TxOutput{out.Value, out.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}
	return txCopy
}

// Verify checks the validity of the signatures for each input in the transaction.
// It reconstructs the data that was signed and uses the public key from the input
// to verify the signature. It returns true if all signatures are valid.
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbaseTx() {
		return true
	}

	for _, in := range tx.Inputs {
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panicf("[TRANSACTION] Verification failed - Invalid parent transaction: %x", in.ID)
		}
	}

	txCopy := tx.TrimmedCopy()

	for inID, in := range tx.Inputs {
		prevTx := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inID].Signature = nil
		txCopy.Inputs[inID].PubKey = prevTx.Outputs[in.Out].PubKeyHash

		dataToVerify := fmt.Sprintf("%x\n", txCopy)

		pubKey := ed25519.PublicKey(in.PubKey)
		if !ed25519.Verify(pubKey, []byte(dataToVerify), in.Signature) {
			return false
		}
		txCopy.Inputs[inID].PubKey = nil
	}

	return true
}

// String provides a human-readable, formatted string representation of the
// transaction. It details the transaction ID, its inputs, and its outputs,
// making it easier to inspect and debug.
func (tx Transaction) String() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("╭─── TRANSACTION [%.16x...] ───╮\n", tx.ID))

	// Inputs
	b.WriteString(fmt.Sprintf("├─ INPUTS (%d)\n", len(tx.Inputs)))
	if len(tx.Inputs) == 0 {
		b.WriteString("│   (No Inputs)\n")
	}
	for i, input := range tx.Inputs {
		if tx.IsCoinbaseTx() {
			b.WriteString(fmt.Sprintf("│   [COINBASE] → Reward Data: %s\n", input.PubKey))
		} else {
			b.WriteString(fmt.Sprintf("│   [%d] From TX: %.16x...\n", i, input.ID))
			b.WriteString(fmt.Sprintf("│       Output Index: %d\n", input.Out))
		}
	}

	// Outputs
	b.WriteString(fmt.Sprintf("├─ OUTPUTS (%d)\n", len(tx.Outputs)))
	if len(tx.Outputs) == 0 {
		b.WriteString("│   (No Outputs)\n")
	}
	for i, output := range tx.Outputs {
		b.WriteString(fmt.Sprintf("│   [%d] To PubKeyHash: %.16x...\n", i, output.PubKeyHash))
		b.WriteString(fmt.Sprintf("│       Value: %d\n", output.Value))
	}

	b.WriteString("╰" + strings.Repeat("─", 50) + "╯")
	return b.String()
}
