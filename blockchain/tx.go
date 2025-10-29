// Package blockchain defines the structures and methods for handling transaction
// inputs and outputs, which are the fundamental components of a transaction.
// These structures manage the flow of value and ownership within the blockchain.
package blockchain

import (
	"bytes"

	"github.com/codila125/foedus-blockchain/wallet"
)

// TxOutput represents a transaction output, which is a specific amount of
// cryptocurrency assigned to a new owner. The output is locked with the public
// key hash of the recipient, ensuring that only the rightful owner can spend it.
type TxOutput struct {
	Value      int    // The amount of cryptocurrency held in this output.
	PubKeyHash []byte // The hash of the public key of the recipient, used to lock the output.
}

// TxInput represents a transaction input, which references a previous unspent
// transaction output (UTXO). It provides proof of ownership through a digital
// signature, allowing the value from the referenced output to be spent.
type TxInput struct {
	ID        []byte // The ID of the transaction containing the output being spent.
	Out       int    // The index of the output within the referenced transaction.
	Signature []byte // A digital signature created with the private key of the owner, proving authorization.
	PubKey    []byte // The public key corresponding to the private key used for the signature.
}

// TxOutputs is a collection of transaction outputs. This structure is primarily
// used for serialization and deserialization purposes, allowing multiple outputs
// to be handled as a single unit.
type TxOutputs struct {
	Outputs []TxOutput
}

// NewTxOutput creates a new transaction output with a specified value, locked to
// a given address. The address is used to derive the public key hash that secures
// the output.
func NewTxOutput(value int, address string) *TxOutput {
	output := &TxOutput{value, nil}
	output.Lock([]byte(address)) // Lock the output to the recipient's address
	return output
}

// Lock sets the public key hash for the transaction output, effectively "locking"
// it to a specific recipient. The address is decoded from Base58, and the public
// key hash is extracted to be used in the output's PubKeyHash field.
func (out *TxOutput) Lock(address []byte) {
	pubKeyHash := wallet.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

// IsLockedWithKey checks if the transaction output is locked with a specific
// public key hash. This method is used to verify that a transaction input has
// the right to spend this output. It returns true if the hashes match.
func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Equal(out.PubKeyHash, pubKeyHash)
}
