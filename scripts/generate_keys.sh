#!/bin/bash
# scripts/generate_keys.sh
# Generates an RSA-2048 key pair for JWT RS256 signing.
# Usage: bash scripts/generate_keys.sh

set -e

mkdir -p certs

echo "Generating RSA-2048 private key..."
openssl genrsa -out certs/private.pem 2048

echo "Extracting public key..."
openssl rsa -in certs/private.pem -pubout -out certs/public.pem

echo "Keys generated:"
echo "  Private: certs/private.pem"
echo "  Public:  certs/public.pem"
echo ""
echo "Add certs/ to .gitignore — NEVER commit private keys to git!"
