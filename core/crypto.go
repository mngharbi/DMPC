package core

import (
	"crypto"
	"crypto/cipher"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"golang.org/x/crypto/chacha20poly1305"
)

/*
	Constants
*/

const (
	AsymmetricKeySizeBits = 2048
	HashingAlgorithm      = crypto.SHA256
	SymmetricKeySize      = chacha20poly1305.KeySize
	SymmetricNonceSize    = chacha20poly1305.NonceSize
	correctChallenge      = "Nizar Gharbi"
)

/*
	Errors
*/
const (
	base64DecodeError      = "Error decoding base64 payload."
	invalidNonceError      = "Invalid nonce provided."
	noSymKeyFoundError     = "No symmetric key passed the challenge."
	payloadDecryptionError = "Payload Decryption failed."
	payloadDecodeError     = "Payload Decoding failed."
)

/*
	Temporary decryption
*/
func (op *TemporaryEncryptedOperation) Decrypt(asymKey *rsa.PrivateKey) (*PermanentEncryptedOperation, error) {
	if op == nil {
		panic("Calling Decrypt on nil pointer.")
	}

	// Base64 decode payload
	payloadBytes, err := base64.StdEncoding.DecodeString(op.Payload)
	if err != nil {
		return nil, errors.New(base64DecodeError)
	}

	// Decrypt payload if encrypted
	var aead cipher.AEAD = nil
	if op.Encryption.Encrypted {

		// Check nonce
		symKeyNonceBytes, err := base64.StdEncoding.DecodeString(op.Encryption.Nonce)
		if err != nil ||
			len(symKeyNonceBytes) != SymmetricNonceSize {
			return nil, errors.New(invalidNonceError)
		}

		// Find a symmetric key that passes challenge
		for symKeyCipher, symKeyChallenge := range op.Encryption.Challenges {
			// Decode symmetric key ciphertext
			symKeyCipherBytes, err := base64.StdEncoding.DecodeString(symKeyCipher)
			if err != nil {
				continue
			}

			// Decrypt symmetric key
			symKeyPlainBytes, symKeyDecryptErr := rsa.DecryptPKCS1v15(nil, asymKey, symKeyCipherBytes)
			if symKeyDecryptErr != nil ||
				len(symKeyPlainBytes) != SymmetricKeySize {
				continue
			}

			// Make symmetric key
			symKeyAead, symKeyAeadErr := chacha20poly1305.New(symKeyPlainBytes)
			if symKeyAeadErr != nil {
				panic("Aead creation failed despite valid key and nonce size")
			}

			// Decode challenge
			symKeyChallengeBytes, err := base64.StdEncoding.DecodeString(symKeyChallenge)
			if err != nil {
				continue
			}

			// Decrypt challenge
			decryptedChallenge, decryptedChallengeErr := symKeyAead.Open(
				symKeyChallengeBytes[:0],
				symKeyNonceBytes,
				symKeyChallengeBytes,
				[]byte{},
			)

			// Test if decrypted challenge is correct
			if decryptedChallengeErr == nil &&
				string(decryptedChallenge) == correctChallenge {
				aead = symKeyAead
				break
			}
		}

		// No symmetric keys worked
		if aead == nil {
			return nil, errors.New(noSymKeyFoundError)
		}

		// Decrypt payload
		payloadDecryptedBytes, payloadDecryptErr := aead.Open(
			payloadBytes[:0],
			symKeyNonceBytes,
			payloadBytes,
			[]byte{},
		)
		if payloadDecryptErr != nil {
			return nil, errors.New(payloadDecryptionError)
		}

		payloadBytes = payloadDecryptedBytes
	}

	// Decode payload
	var decodedOp PermanentEncryptedOperation
	payloadDecodeErr := decodedOp.Decode(payloadBytes)
	if payloadDecodeErr != nil {
		return nil, errors.New(payloadDecodeError)
	}

	return &decodedOp, nil
}
