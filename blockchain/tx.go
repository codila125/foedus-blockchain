package blockchain

import (
	"bytes"
	"encoding/gob"
	"github.com/codila125/foedus-blockchain/wallet"
)

type TxOutput struct {
	Value      int    // Value of the output
	PubKeyHash []byte // Hash of the public key that can unlock this output, i.e., the recipient's address
}

type TxInput struct {
	ID        []byte // ID of the transaction containing the output to be used as input
	Out       int    // Index of the output in the transaction
	Signature []byte // Signature of the input
	PubKey    []byte // Public key of the input
}

type TxOutputs struct {
	Outputs []TxOutput // Slice of transaction outputs
}

func NewTxOutput(value int, address string) *TxOutput {
	/*
		Creates a new transaction output.
		'value' is the amount of currency being sent.
		'address' is the recipient's address.
	*/
	output := &TxOutput{value, nil}
	output.Lock([]byte(address)) // Lock the output to the recipient's address
	return output
}

func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := wallet.PublicKeyHash(in.PubKey)

	return bytes.Equal(lockingHash, pubKeyHash)
}

func (out *TxOutput) Lock(address []byte) {
	/*
		Locks the output to a specific address by setting the PubKeyHash.
		'address' is the recipient's address.
	*/
	pubKeyHash := wallet.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4] // Remove version byte and checksum
	out.PubKeyHash = pubKeyHash
}

func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	/*
		Checks if the output can be unlocked with the given public key hash.
		'pubKeyHash' is the hash of the public key to check against.
	*/
	return bytes.Equal(out.PubKeyHash, pubKeyHash)
}

func (outs TxOutputs) Serialize() []byte {
	/*
		Serializes the TxOutputs struct into a byte slice.
	*/
	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(outs)
	Handle(err)
	return buff.Bytes()
}

func DeserializeOutputs(data []byte) TxOutputs {
	/*
		Deserializes a byte slice into a TxOutputs struct.
		'data' is the byte slice to be deserialized.
	*/
	var outputs TxOutputs
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&outputs)
	Handle(err)
	return outputs
}
