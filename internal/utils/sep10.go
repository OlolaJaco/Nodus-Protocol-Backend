package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
	"github.com/stellar/go/txnbuild"
)

// Sep10Manager handles SEP-10 Stellar Web Authentication challenge issuance and verification.
type Sep10Manager struct {
	serverKP      *keypair.Full
	webAuthDomain string
	passphrase    string
	challengeTTL  time.Duration
}

// NewSep10Manager parses the server secret key and prepares the SEP-10 manager.
// stellarNetwork must be "testnet" or "mainnet".
func NewSep10Manager(secretKey, webAuthDomain, stellarNetwork string, ttl time.Duration) (*Sep10Manager, error) {
	kp, err := keypair.ParseFull(secretKey)
	if err != nil {
		return nil, fmt.Errorf("invalid stellar server secret key: %w", err)
	}

	passphrase := network.TestNetworkPassphrase
	if stellarNetwork == "mainnet" {
		passphrase = network.PublicNetworkPassphrase
	}

	return &Sep10Manager{
		serverKP:      kp,
		webAuthDomain: webAuthDomain,
		passphrase:    passphrase,
		challengeTTL:  ttl,
	}, nil
}

// ServerAddress returns the server's Stellar account address (G...).
func (m *Sep10Manager) ServerAddress() string {
	return m.serverKP.Address()
}

// BuildChallenge creates a SEP-10 challenge transaction for the given Stellar account ID.
// Returns a base64-encoded XDR transaction envelope that the client must sign and return.
func (m *Sep10Manager) BuildChallenge(accountID string) (string, error) {
	if _, err := keypair.ParseAddress(accountID); err != nil {
		return "", errors.New("invalid stellar account id")
	}

	nonce := make([]byte, 32)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("nonce generation failed: %w", err)
	}

	now := time.Now().Unix()

	tx, err := txnbuild.NewTransaction(txnbuild.TransactionParams{
		SourceAccount: &txnbuild.SimpleAccount{
			AccountID: m.serverKP.Address(),
			Sequence:  0,
		},
		IncrementSequenceNum: false,
		Operations: []txnbuild.Operation{
			// First op: client account as source — client must sign as this account.
			&txnbuild.ManageData{
				Name:          m.webAuthDomain + " auth",
				Value:         []byte(base64.StdEncoding.EncodeToString(nonce)),
				SourceAccount: accountID,
			},
			// Second op: server declares the web_auth_domain.
			&txnbuild.ManageData{
				Name:          "web_auth_domain",
				Value:         []byte(m.webAuthDomain),
				SourceAccount: m.serverKP.Address(),
			},
		},
		BaseFee: txnbuild.MinBaseFee,
		Preconditions: txnbuild.Preconditions{
			TimeBounds: txnbuild.NewTimebounds(now, now+int64(m.challengeTTL.Seconds())),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to build challenge transaction: %w", err)
	}

	tx, err = tx.Sign(m.passphrase, m.serverKP)
	if err != nil {
		return "", fmt.Errorf("failed to sign challenge: %w", err)
	}

	b64, err := tx.Base64()
	if err != nil {
		return "", fmt.Errorf("failed to encode challenge: %w", err)
	}

	return b64, nil
}

// VerifyChallenge validates a client-signed SEP-10 challenge and returns the authenticated
// Stellar account ID. It checks time bounds, server signature, and client signature.
func (m *Sep10Manager) VerifyChallenge(xdrBase64 string) (string, error) {
	// ReadChallengeTx validates time bounds and server signature, and extracts the client account.
	_, clientAccountID, _, _, err := txnbuild.ReadChallengeTx(
		xdrBase64,
		m.serverKP.Address(),
		m.passphrase,
		m.webAuthDomain,
		[]string{m.webAuthDomain},
	)
	if err != nil {
		return "", fmt.Errorf("invalid challenge transaction: %w", err)
	}

	// VerifyChallengeTxSigners confirms the client account actually signed the challenge.
	_, err = txnbuild.VerifyChallengeTxSigners(
		xdrBase64,
		m.serverKP.Address(),
		m.passphrase,
		m.webAuthDomain,
		[]string{m.webAuthDomain},
		clientAccountID,
	)
	if err != nil {
		return "", fmt.Errorf("challenge signature verification failed: %w", err)
	}

	return clientAccountID, nil
}
