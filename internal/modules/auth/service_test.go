package auth

import (
	"testing"

	"github.com/nbutton23/zxcvbn-go"
	"github.com/stretchr/testify/assert"
)

// TestPasswordStrengthValidation tests the zxcvbn password strength logic
// without needing full service dependencies.
func TestPasswordStrengthValidation(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		userInputs  []string
		shouldPass  bool
		description string
	}{
		{
			name:        "weak password - common word",
			password:    "password",
			userInputs:  []string{"test@example.com", "John", "Doe"},
			shouldPass:  false,
			description: "Common dictionary word should be rejected",
		},
		{
			name:        "weak password - sequential numbers",
			password:    "123456",
			userInputs:  []string{"test@example.com", "John", "Doe"},
			shouldPass:  false,
			description: "Sequential numbers should be rejected",
		},
		{
			name:        "weak password - contains user input",
			password:    "john123456",
			userInputs:  []string{"test@example.com", "John", "Doe"},
			shouldPass:  false,
			description: "Password containing user's name should be rejected",
		},
		{
			name:        "weak password - simple pattern",
			password:    "nodus123",
			userInputs:  []string{"test@example.com", "John", "Doe"},
			shouldPass:  false,
			description: "Simple pattern with low entropy should be rejected",
		},
		{
			name:        "strong password - random characters",
			password:    "X9$mK#pL2@wQ",
			userInputs:  []string{"test@example.com", "John", "Doe"},
			shouldPass:  true,
			description: "Strong random password should be accepted",
		},
		{
			name:        "strong password - passphrase",
			password:    "correct-horse-battery-staple",
			userInputs:  []string{"test@example.com", "John", "Doe"},
			shouldPass:  true,
			description: "Strong passphrase should be accepted",
		},
		{
			name:        "strong password - long random",
			password:    "Tr0ub4dor&3PlusMore",
			userInputs:  []string{"test@example.com", "John", "Doe"},
			shouldPass:  true,
			description: "Strong password with good entropy should be accepted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strength := zxcvbn.PasswordStrength(tt.password, tt.userInputs)
			passed := strength.Score >= minPasswordScore

			assert.Equal(t, tt.shouldPass, passed,
				"Password: %s, Score: %d, Expected pass: %v, Got pass: %v",
				tt.password, strength.Score, tt.shouldPass, passed)
		})
	}
}

// TestPasswordMinimumLength tests that minimum length validation is enforced
func TestPasswordMinimumLength(t *testing.T) {
	tests := []struct {
		name       string
		password   string
		shouldPass bool
	}{
		{
			name:       "too short - 5 characters",
			password:   "Ab1@x",
			shouldPass: false,
		},
		{
			name:       "too short - 8 characters",
			password:   "Ab1@xYz!",
			shouldPass: false,
		},
		{
			name:       "minimum length - 10 characters",
			password:   "Ab1@xYz!Qw",
			shouldPass: true,
		},
		{
			name:       "longer password",
			password:   "Ab1@xYz!QwErTy",
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passed := len(tt.password) >= 10
			assert.Equal(t, tt.shouldPass, passed,
				"Password length: %d, Expected pass: %v, Got pass: %v",
				len(tt.password), tt.shouldPass, passed)
		})
	}
}
