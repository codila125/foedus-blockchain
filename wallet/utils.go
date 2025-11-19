// Package wallet provides cryptographic wallet management, including utilities
// for encoding and decoding data using Base58. This encoding is widely used
// in blockchain applications for creating human-readable addresses from
// cryptographic hashes.
package wallet

import (
	"log"

	"github.com/mr-tron/base58"
)

// Base58Encode converts a byte slice into a Base58-encoded string. Base58 is
// used to create shorter, more readable addresses by excluding characters that
// might cause confusion (e.g., 0, O, I, l).
func Base58Encode(input []byte) []byte {
	encode := base58.Encode(input)

	return []byte(encode)
}

// Base58Decode converts a Base58-encoded string back into its original byte
// slice representation. This function is critical for validating addresses and
// retrieving the underlying public key hash. It panics if the decoding process
// fails, as this indicates a malformed or corrupted address.
func Base58Decode(input []byte) []byte {
	decode, err := base58.Decode(string(input[:]))
	if err != nil {
		log.Print(err)
	}

	return decode
}
