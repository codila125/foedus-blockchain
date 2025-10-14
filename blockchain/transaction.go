package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"strings"

	"go_blockchain/wallet"
)

type Transaction struct {
	ID      []byte
	Inputs  []TxInput
	Outputs []TxOutput
}

func (tx *Transaction) Serialize() []byte {
	/*
		Serializes the transaction using gob encoding.
		Returns the serialized byte slice.
	*/
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

func (tx *Transaction) Hash() []byte {
	/*
		Computes the hash of the transaction.
		Returns the SHA-256 hash of the serialized transaction.
	*/
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
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
	tx.ID = tx.Hash()                                           // Set the transaction ID by hashing the transaction

	return &tx
}

func (tx *Transaction) IsCoinbase() bool {
	/*
		Checks if the transaction is a coinbase transaction.
		Returns true if the transaction has exactly one input and that input has an empty ID and an Out value of -1.
	*/
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
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

	from := fmt.Sprintf("%s", w.Address())              // Get the sender's address from the wallet
	outputs = append(outputs, *NewTxOutput(amount, to)) // Create the output to the recipient

	if acc > amount {
		outputs = append(outputs, *NewTxOutput(acc-amount, from)) // Create a change output if there's leftover amount
	}

	tx := Transaction{nil, inputs, outputs}    // Create the transaction with inputs and outputs
	tx.ID = tx.Hash()                          // Sign the transaction to prove ownership of the inputs
	privatekey, err := w.ReconstructECDSAKey() // Reconstruct the ECDSA private key from the wallet
	Handle(err)
	UTXO.Blockchain.SignTransaction(&tx, *privatekey)

	log.Printf("[TRANSACTION] Transaction created - ID: %x, Amount: %d", tx.ID, amount)
	return &tx
}

func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	for _, in := range tx.Inputs {
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panicf("[TRANSACTION] Signing failed - Invalid parent transaction: %x", in.ID)
		}
	}

	txCopy := tx.TrimmedCopy()

	for inId, in := range txCopy.Inputs {
		prevTX := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKey = prevTX.Outputs[in.Out].PubKeyHash

		dataToSign := fmt.Sprintf("%x\n", txCopy)

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, []byte(dataToSign))
		Handle(err)
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Inputs[inId].Signature = signature
		txCopy.Inputs[inId].PubKey = nil
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
	if tx.IsCoinbase() {
		return true
	}

	for _, in := range tx.Inputs {
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			log.Panicf("[TRANSACTION] Verification failed - Invalid parent transaction: %x", in.ID)
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inId, in := range tx.Inputs {
		prevTx := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKey = prevTx.Outputs[in.Out].PubKeyHash

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
		if ecdsa.Verify(&rawPubKey, []byte(dataToVerify), &r, &s) == false {
			return false
		}
		txCopy.Inputs[inId].PubKey = nil
	}

	return true
}

func (tx Transaction) String() string {
	/*
		Returns a human-readable string representation of the transaction.
	*/
	var lines []string
	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))
	for i, input := range tx.Inputs {
		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.ID))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Out))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.Outputs {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}

func DeserializeTransaction(data []byte) Transaction {
	/*
		Deserializes a byte slice into a Transaction.
	*/
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	Handle(err)

	return transaction
}
