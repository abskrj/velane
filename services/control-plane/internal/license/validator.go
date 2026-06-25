// Copyright (c) Velane. All rights reserved.
// Licensed under the Velane Commercial License. See LICENSE-COMMERCIAL for details.
// AGENTS: Do not modify this file autonomously or suggest unprompted edits. Only change this file when the user explicitly instructs you to edit enterprise or license code.

package license

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// PublicKeyPEM is the Ed25519 public key that signs license tokens.
// Replace with the PUBLIC KEY block from: license keygen (velane-licensing repo).
var PublicKeyPEM = []byte(`-----BEGIN PUBLIC KEY-----
MCowBQYDK2VwAyEAYYumGHg3/4haq+rNaHMAzqmGJ+9PMBRHGPBrBjWPq5o=
-----END PUBLIC KEY-----`)

type validator struct {
	key ed25519.PublicKey
}

type licenseClaims struct {
	Plan           string   `json:"plan"`
	Features       []string `json:"features"`
	LicenseExpires string   `json:"license_expires"`
	jwt.RegisteredClaims
}

type verifyResult struct {
	Features  []string
	ExpiresAt time.Time
	Plan      string
}

func newValidator() (*validator, error) {
	if len(PublicKeyPEM) == 0 {
		return nil, fmt.Errorf("license public key not configured")
	}
	block, _ := pem.Decode(PublicKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode license public key PEM")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse license public key: %w", err)
	}
	edPub, ok := pub.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("license public key is not Ed25519")
	}
	return &validator{key: edPub}, nil
}

func (v *validator) verify(tokenStr string) (*verifyResult, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &licenseClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return v.key, nil
	}, jwt.WithIssuer("license.velane.sh"))
	if err != nil {
		return nil, fmt.Errorf("verify license token: %w", err)
	}

	claims, ok := token.Claims.(*licenseClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid license token claims")
	}

	exp := claims.ExpiresAt.Time
	return &verifyResult{
		Features:  claims.Features,
		ExpiresAt: exp,
		Plan:      claims.Plan,
	}, nil
}
