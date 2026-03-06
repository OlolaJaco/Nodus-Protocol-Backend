package utils

import (
	"errors"
	"strings"
)

var (
	ErrInvalidStellarAddress = errors.New("stellar address must be 56 characters and start with G")
	ErrInvalidContractID     = errors.New("stellar contract ID must be 56 characters and start with C")
)

// ValidateStellarAddress checks that s is a valid Stellar account public key
// (StrKey-encoded Ed25519 public key, 56 chars, starts with G).
func ValidateStellarAddress(s string) error {
	s = strings.TrimSpace(s)
	if len(s) != 56 || !strings.HasPrefix(s, "G") {
		return ErrInvalidStellarAddress
	}
	return nil
}

// ValidateContractID checks that s looks like a Stellar contract address
// (56 chars, starts with C).
func ValidateContractID(s string) error {
	s = strings.TrimSpace(s)
	if len(s) != 56 || !strings.HasPrefix(s, "C") {
		return ErrInvalidContractID
	}
	return nil
}

// ShortenAddress returns a display-friendly version of a Stellar address:
// "GAAAA…ZZZZ".
func ShortenAddress(addr string) string {
	if len(addr) < 10 {
		return addr
	}
	return addr[:5] + "…" + addr[len(addr)-4:]
}

// IsStellarAddress returns true if the string passes ValidateStellarAddress.
func IsStellarAddress(s string) bool {
	return ValidateStellarAddress(s) == nil
}
