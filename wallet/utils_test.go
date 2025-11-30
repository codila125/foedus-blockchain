package wallet

import (
	"bytes"
	"testing"
)

// go test -run TestBase58
// go test -bench=BenchmarkBase58

// TestBase58Encode tests the Base58 encoding function with various inputs.
func TestBase58Encode(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "empty input",
			input:    []byte{},
			expected: "",
		},
		{
			name:     "single byte zero",
			input:    []byte{0},
			expected: "1",
		},
		{
			name:     "hello world",
			input:    []byte("hello world"),
			expected: "StV1DL6CwTryKyV",
		},
		{
			name:     "all zeros",
			input:    []byte{0, 0, 0},
			expected: "111",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Base58Encode(tt.input)
			if string(result) != tt.expected {
				t.Errorf("Base58Encode(%v) = %s, want %s", tt.input, string(result), tt.expected)
			}
		})
	}
}

// TestBase58Decode tests the Base58 decoding function with various inputs.
func TestBase58Decode(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "single 1 (zero)",
			input:    []byte("1"),
			expected: []byte{0},
		},
		{
			name:     "hello world encoded",
			input:    []byte("StV1DL6CwTryKyV"),
			expected: []byte("hello world"),
		},
		{
			name:     "multiple ones (zeros)",
			input:    []byte("111"),
			expected: []byte{0, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Base58Decode(tt.input)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("Base58Decode(%s) = %v, want %v", string(tt.input), result, tt.expected)
			}
		})
	}
}

// TestBase58RoundTrip tests that encoding then decoding returns the original data.
func TestBase58RoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{
			name:  "empty",
			input: []byte{},
		},
		{
			name:  "single byte",
			input: []byte{42},
		},
		{
			name:  "text data",
			input: []byte("blockchain wallet test"),
		},
		{
			name:  "binary data",
			input: []byte{0x00, 0xff, 0x10, 0xab, 0xcd},
		},
		{
			name:  "256 bytes",
			input: make([]byte, 256),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := Base58Encode(tt.input)
			decoded := Base58Decode(encoded)
			if !bytes.Equal(decoded, tt.input) {
				t.Errorf("RoundTrip failed: got %v, want %v", decoded, tt.input)
			}
		})
	}
}

// Benchmark tests for utils

func BenchmarkBase58Encode(b *testing.B) {
	data := make([]byte, 32)
	for i := range data {
		data[i] = byte(i)
	}

	for b.Loop() {
		Base58Encode(data)
	}
}

func BenchmarkBase58Decode(b *testing.B) {
	data := make([]byte, 32)
	for i := range data {
		data[i] = byte(i)
	}
	encoded := Base58Encode(data)

	for b.Loop() {
		Base58Decode(encoded)
	}
}
