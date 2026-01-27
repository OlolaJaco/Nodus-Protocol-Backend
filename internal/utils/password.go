package utils

import (
	"math"
	"unicode"
)

// PasswordStrength returns a 0–100 score for the given password and a
// list of unmet requirements. Passwords below 40 should be rejected.
func PasswordStrength(password string) (score int, issues []string) {
	length := len(password)

	var hasUpper, hasLower, hasDigit, hasSymbol bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case unicode.IsPunct(r) || unicode.IsSymbol(r):
			hasSymbol = true
		}
	}

	if length < 8 {
		issues = append(issues, "must be at least 8 characters")
	}
	if !hasUpper {
		issues = append(issues, "must contain an uppercase letter")
	}
	if !hasLower {
		issues = append(issues, "must contain a lowercase letter")
	}
	if !hasDigit {
		issues = append(issues, "must contain a digit")
	}

	// Entropy-based score (Shannon entropy × character-set bonus)
	entropy := shannonEntropy(password)
	entropyScore := int(math.Min(entropy*10, 60))

	bonuses := 0
	if length >= 12 {
		bonuses += 10
	}
	if length >= 16 {
		bonuses += 10
	}
	if hasSymbol {
		bonuses += 10
	}
	if hasUpper && hasLower && hasDigit {
		bonuses += 10
	}

	score = entropyScore + bonuses
	if score > 100 {
		score = 100
	}
	return score, issues
}

func shannonEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}
	freq := make(map[rune]float64)
	for _, r := range s {
		freq[r]++
	}
	n := float64(len(s))
	var entropy float64
	for _, count := range freq {
		p := count / n
		entropy -= p * math.Log2(p)
	}
	return entropy
}
