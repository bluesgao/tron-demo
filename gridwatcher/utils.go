package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/btcsuite/btcutil/base58"
)

// TronBase58ToEvmHex converts Tron base58 (T...) to EVM style hex (0x...., 40 hex chars, lower).
// Tron address raw bytes: 21 bytes, first byte is 0x41, followed by 20 bytes address.
func TronBase58ToEvmHex(addr string) (string, error) {
	raw := base58.Decode(addr)
	if len(raw) < 4+21 {
		return "", fmt.Errorf("invalid tron base58 length")
	}
	// payload = 21 bytes, checksum = last 4 bytes
	payload := raw[:len(raw)-4]
	checksum := raw[len(raw)-4:]

	// verify checksum: first 4 bytes of sha256(sha256(payload))
	h1 := sha256.Sum256(payload)
	h2 := sha256.Sum256(h1[:])
	if !equal4(checksum, h2[:4]) {
		return "", fmt.Errorf("invalid checksum")
	}
	if len(payload) != 21 || payload[0] != 0x41 {
		return "", fmt.Errorf("invalid tron payload prefix")
	}
	eth20 := payload[1:] // 20 bytes
	return "0x" + strings.ToLower(hex.EncodeToString(eth20)), nil
}

func equal4(a, b []byte) bool {
	if len(a) != 4 || len(b) != 4 {
		return false
	}
	for i := 0; i < 4; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
