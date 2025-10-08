// Package wallet provides utilities for managing cryptocurrency wallets.
package wallet

import (
	"log"

	"github.com/mr-tron/base58"
)

func Base58Encode(input []byte) []byte {
	/*
		Encodes the input byte slice using Base58 encoding.
		Returns the Base58 encoded byte slice.
	*/
	encode := base58.Encode(input)

	return []byte(encode)
}

func Base58Decode(input []byte) []byte {
	/*
		Decodes the Base58 encoded byte slice back to its original byte slice.
		Returns the decoded byte slice.
	*/
	decode, err := base58.Decode(string(input[:]))
	if err != nil {
		log.Panic(err)
	}

	return decode
}
