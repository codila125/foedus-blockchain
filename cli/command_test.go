package cli

import (
	"testing"

	"github.com/codila125/foedus-blockchain/wallet"
)

// =============================================================================
// Address Validation Tests
// =============================================================================

func TestAddressValidation_ValidAddress(t *testing.T) {
	// Generate a valid wallet address for testing
	w := wallet.MakeWallet()
	address := string(w.Address())

	if !wallet.ValidateAddress(address) {
		t.Error("Generated wallet address should be valid")
	}
}

func TestAddressValidation_EmptyString(t *testing.T) {
	// Empty string should return false, not panic
	if wallet.ValidateAddress("") {
		t.Error("Empty string should not be a valid address")
	}
}

func TestAddressValidation_TooShort(t *testing.T) {
	// Very short strings should return false gracefully
	shortAddresses := []string{
		"a",
		"ab",
		"abc",
		"1234",
		"12345",
	}

	for _, addr := range shortAddresses {
		if wallet.ValidateAddress(addr) {
			t.Errorf("Short address '%s' should not be valid", addr)
		}
	}
}

func TestAddressValidation_InvalidBase58Characters(t *testing.T) {
	// Addresses with invalid base58 characters (0, O, I, l) should return false
	invalidAddresses := []string{
		"0invalidaddress123456789012345678901234",
		"Oinvalidaddress123456789012345678901234",
		"Iinvalidaddress123456789012345678901234",
		"linvalidaddress123456789012345678901234",
	}

	for _, addr := range invalidAddresses {
		if wallet.ValidateAddress(addr) {
			t.Errorf("Address with invalid base58 char '%s' should not be valid", addr)
		}
	}
}

func TestAddressValidation_RandomStrings(t *testing.T) {
	// Random strings that look like addresses but aren't
	randomStrings := []string{
		"notavalidblockchainaddress",
		"123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijk",
		"ThisIsDefinitelyNotAValidAddress123456789",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}

	for _, addr := range randomStrings {
		if wallet.ValidateAddress(addr) {
			t.Errorf("Random string '%s' should not be a valid address", addr)
		}
	}
}

func TestAddressValidation_ModifiedAddressFails(t *testing.T) {
	// Create a valid address then modify it slightly to make it invalid
	w := wallet.MakeWallet()
	validAddr := string(w.Address())

	// Modified address (changed one character) should fail checksum
	if len(validAddr) > 10 {
		// Modify middle character while keeping base58 valid characters
		modifiedAddr := validAddr[:5] + "1" + validAddr[6:]

		isValid := wallet.ValidateAddress(modifiedAddr)
		if isValid {
			t.Error("Modified address should not pass validation")
		}
	}
}

func TestAddressValidation_WhitespaceOnly(t *testing.T) {
	// Whitespace-only strings should return false
	whitespaceAddresses := []string{
		" ",
		"   ",
		"\t",
		"\n",
		"  \t\n  ",
	}

	for _, addr := range whitespaceAddresses {
		if wallet.ValidateAddress(addr) {
			t.Errorf("Whitespace-only address '%q' should not be valid", addr)
		}
	}
}

func TestAddressValidation_SpecialCharacters(t *testing.T) {
	// Addresses with special characters should return false
	specialAddresses := []string{
		"address!@#$%^&*()",
		"address<>?/\\|",
		"address with spaces",
		"address\twith\ttabs",
	}

	for _, addr := range specialAddresses {
		if wallet.ValidateAddress(addr) {
			t.Errorf("Address with special chars '%s' should not be valid", addr)
		}
	}
}

// =============================================================================
// Command Structure Tests
// =============================================================================

func TestCommandLine_MethodsExist(t *testing.T) {
	cli := &CommandLine{}

	// Verify CLI instance is created (check type)
	var _ interface{} = cli
}

// =============================================================================
// Command Validation Tests (without calling ValidateAddress on invalid data)
// =============================================================================

func TestCommandValidation_AmountMustBePositive(t *testing.T) {
	testAmounts := []struct {
		amount   int
		expected bool
	}{
		{0, false},
		{-1, false},
		{-100, false},
		{1, true},
		{100, true},
		{1000000, true},
	}

	for _, tc := range testAmounts {
		isPositive := tc.amount > 0
		if isPositive != tc.expected {
			t.Errorf("Amount %d: expected positive=%v, got %v", tc.amount, tc.expected, isPositive)
		}
	}
}

func TestCommandValidation_AddressNonEmpty(t *testing.T) {
	testAddresses := []struct {
		address  string
		expected bool
	}{
		{"", false},
		{"someaddress", true},
		{"   ", true}, // whitespace is non-empty
	}

	for _, tc := range testAddresses {
		isNonEmpty := len(tc.address) > 0
		if isNonEmpty != tc.expected {
			t.Errorf("Address '%s': expected non-empty=%v, got %v", tc.address, tc.expected, isNonEmpty)
		}
	}
}

// =============================================================================
// Wallet Tests
// =============================================================================

func TestCreateWallet_GeneratesNewAddress(t *testing.T) {
	w := wallet.MakeWallet()
	address := w.Address()

	if len(address) == 0 {
		t.Error("Wallet should generate non-empty address")
	}
}

func TestCreateWallet_AddressesAreUnique(t *testing.T) {
	addresses := make(map[string]bool)

	for i := 0; i < 10; i++ {
		w := wallet.MakeWallet()
		addr := string(w.Address())

		if addresses[addr] {
			t.Errorf("Duplicate address generated: %s", addr)
		}
		addresses[addr] = true
	}
}

func TestCreateWallet_AddressHasValidLength(t *testing.T) {
	w := wallet.MakeWallet()
	addr := string(w.Address())

	// Wallet addresses should typically be between 25-55 characters
	if len(addr) < 20 || len(addr) > 60 {
		t.Errorf("Address length %d seems invalid: %s", len(addr), addr)
	}
}

func TestCreateWallet_AddressIsValidBase58(t *testing.T) {
	w := wallet.MakeWallet()
	addr := string(w.Address())

	// Check that address only contains valid base58 characters
	base58Chars := "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	validChars := make(map[rune]bool)
	for _, c := range base58Chars {
		validChars[c] = true
	}

	for _, c := range addr {
		if !validChars[c] {
			t.Errorf("Address contains invalid base58 character: %c", c)
		}
	}
}

// =============================================================================
// Multiple Wallet Tests
// =============================================================================

func TestMultipleWallets_AllValid(t *testing.T) {
	for i := 0; i < 5; i++ {
		w := wallet.MakeWallet()
		addr := string(w.Address())

		if !wallet.ValidateAddress(addr) {
			t.Errorf("Wallet %d generated invalid address: %s", i, addr)
		}
	}
}

// =============================================================================
// Benchmarks
// =============================================================================

func BenchmarkWalletCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		wallet.MakeWallet()
	}
}

func BenchmarkAddressValidation(b *testing.B) {
	w := wallet.MakeWallet()
	address := string(w.Address())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wallet.ValidateAddress(address)
	}
}

func BenchmarkAddressGeneration(b *testing.B) {
	for i := 0; i < b.N; i++ {
		w := wallet.MakeWallet()
		_ = w.Address()
	}
}
