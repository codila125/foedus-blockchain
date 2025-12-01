package blockchain

import (
	"github.com/codila125/foedus-blockchain/wallet"
)

// TestHelper provides common test utilities for blockchain tests
type TestHelper struct {
	Wallet  *wallet.Wallet
	Address string
}

// NewTestHelper creates a new test helper with a generated wallet
func NewTestHelper() *TestHelper {
	w := wallet.MakeWallet()
	return &TestHelper{
		Wallet:  w,
		Address: string(w.Address()),
	}
}

// CreateTestWallet creates a new wallet and returns its address
func CreateTestWallet() string {
	w := wallet.MakeWallet()
	return string(w.Address())
}

// CreateTestCoinbaseTx creates a coinbase transaction with a valid address
func CreateTestCoinbaseTx(data string) *Transaction {
	address := CreateTestWallet()
	return CoinbaseTx(address, data)
}

// CreateTestCoinbaseOp creates a coinbase contract with a valid address
func CreateTestCoinbaseOp(data string) *Contract {
	address := CreateTestWallet()
	return CoinbaseOp(address, data)
}
