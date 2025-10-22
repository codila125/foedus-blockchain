package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/codila125/foedus-blockchain/wallet"
)

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

func CoinbaseTx(to, data string) *Transaction {
	/*
		Creates a coinbase transaction that rewards the miner.
		'to' is the address to send the reward to.
		'data' is arbitrary data, often used to include a message or extra information.
	*/
	if data == "" {
		randData := make([]byte, 20) // Generate 20 random bytes
		_, err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}
		data = fmt.Sprintf("%x", randData) // Convert random bytes to a hex string
	}

	txin := TxInput{[]byte{}, -1, nil, []byte(data)} // Coinbase input has no previous transaction, hence ID is empty and Out is -1
	txout := NewTxOutput(20, to)                     // Coinbase transaction typically has a fixed reward, here set to 20

	tx := Transaction{nil, []TxInput{txin}, []TxOutput{*txout}} // Create the transaction with the input and output
	tx.ID = tx.HashTransaction()                                // Set the transaction ID by hashing the transaction

	return &tx
}

func (tx *Transaction) IsCoinbaseTx() bool {
	/*
		Checks if the transaction is a coinbase transaction.
		Returns true if the transaction has exactly one input and that input has an empty ID and an Out value of -1.
	*/
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

func (blockchain *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	/*
		Finds and returns a transaction by its ID by scanning through all blocks in the blockchain.
		Returns the transaction if found, otherwise returns an error.
	*/
	iterator := blockchain.Iterator()
	// Iterate through the blocks in the blockchain
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

	return Transaction{}, errors.New("Transaction does not exist")
}

func (blockchain *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	/*
		Signs a transaction using the provided private key.
		'tx' is the transaction to be signed.
		'privKey' is the ECDSA private key used for signing.
	*/
	if tx.IsCoinbaseTx() {
		return
	}
	prevTXs := make(map[string]Transaction)
	for _, in := range tx.Inputs {
		prevTX, err := blockchain.FindTransaction(in.ID) // Find the previous transaction referenced by the input
		Handle(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	tx.Sign(privKey, prevTXs) // Sign the transaction with the private key and previous transactions
}

func (blockchain *BlockChain) VerifyTransaction(tx *Transaction, txMap map[string]Transaction) bool {
	/*
	   Verifies the signatures of a transaction.
	   'tx' is the transaction to be verified.
	   'txMap' is a map of other transactions in the same block/pool.
	   Returns true if the transaction is valid, false otherwise.
	*/
	if tx.IsCoinbaseTx() {
		return true
	}
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		// First, check if the previous transaction is in the current pool of transactions.
		if prevTx, ok := txMap[hex.EncodeToString(in.ID)]; ok {
			prevTXs[hex.EncodeToString(prevTx.ID)] = prevTx
		} else {
			// If not in the pool, search the blockchain.
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

func NewTransaction(w *wallet.Wallet, to string, amount int, UTXO *UTXOSet) *Transaction {
	/*
		Creates a new transaction from one address to another.
		'from' is the sender's address.
		'to' is the recipient's address.
		'amount' is the amount to send.
		'UTXO' is the UTXO set used to find spendable outputs.
	*/
	var inputs []TxInput
	var outputs []TxOutput
	// Get the wallet for the sender's address
	pubKeyHash := wallet.PublicKeyHash(w.PublicKey)                    // Get the public key hash from the wallet's public key
	acc, validOutputs := UTXO.FindSpendableOutputs(pubKeyHash, amount) // Find spendable outputs for the public key hash

	// Check if the accumulated amount is less than the requested amount
	if acc < amount {
		log.Panicf("[TRANSACTION] Insufficient funds - Required: %d, Available: %d", amount, acc)
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		Handle(err)

		// Create a new input for each output which is being used
		for _, out := range outs {
			input := TxInput{txID, out, nil, w.PublicKey} // Create a new input referencing the output
			inputs = append(inputs, input)
		}
	}

	from := string(w.Address())                         // Get the sender's address from the wallet
	outputs = append(outputs, *NewTxOutput(amount, to)) // Create the output to the recipient

	if acc > amount {
		outputs = append(outputs, *NewTxOutput(acc-amount, from)) // Create a change output if there's leftover amount
	}

	tx := Transaction{nil, inputs, outputs}    // Create the transaction with inputs and outputs
	tx.ID = tx.HashTransaction()               // Sign the transaction to prove ownership of the inputs
	privatekey, err := w.ReconstructECDSAKey() // Reconstruct the ECDSA private key from the wallet
	Handle(err)
	UTXO.Blockchain.SignTransaction(&tx, *privatekey)

	log.Printf("[TRANSACTION] Transaction created - ID: %x, Amount: %d", tx.ID, amount)
	return &tx
}

func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
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

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, []byte(dataToSign))
		Handle(err)
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Inputs[inID].Signature = signature
		txCopy.Inputs[inID].PubKey = nil
	}
}

func (tx *Transaction) TrimmedCopy() Transaction {
	/*
		Creates a trimmed copy of the transaction with empty signatures and public keys in the inputs.
		Returns the trimmed copy of the transaction.
	*/
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
	curve := elliptic.P256()

	for inID, in := range tx.Inputs {
		prevTx := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inID].Signature = nil
		txCopy.Inputs[inID].PubKey = prevTx.Outputs[in.Out].PubKeyHash

		r := big.Int{}
		s := big.Int{}

		sigLen := len(in.Signature)
		r.SetBytes(in.Signature[:(sigLen / 2)])
		s.SetBytes(in.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(in.PubKey)
		x.SetBytes(in.PubKey[:(keyLen / 2)])
		y.SetBytes(in.PubKey[(keyLen / 2):])

		dataToVerify := fmt.Sprintf("%x\n", txCopy)

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if !ecdsa.Verify(&rawPubKey, []byte(dataToVerify), &r, &s) {
			return false
		}
		txCopy.Inputs[inID].PubKey = nil
	}

	return true
}

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
