package core

import (
	"crypto"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"golang.org/x/crypto/chacha20poly1305"
	"io"
)

/*
	Constants
*/

const (
	AsymmetricKeySizeBits         = 2048
	AsymmetricKeySizeBytes        = 256
	maxAsymmetricCiphertextLength = AsymmetricKeySizeBytes - 11
	HashingAlgorithm              = crypto.SHA256
	SymmetricKeySize              = chacha20poly1305.KeySize
	SymmetricNonceSize            = chacha20poly1305.NonceSize
	CorrectChallenge              = "Nizar Gharbi"
)

/*
	Errors
*/
var (
	base64DecodeError              error = errors.New("Error decoding base64.")
	invalidNonceError              error = errors.New("Invalid nonce provided.")
	invalidSymmetricKeyError       error = errors.New("Invalid key provided.")
	aeadCreationError              error = errors.New("Aead creation failed.")
	noSymmetricKeyFoundError       error = errors.New("No symmetric key passed the challenge.")
	signError                      error = errors.New("Signing failed.")
	asymmetrictEncryptionError     error = errors.New("Asymmetric encryption failed.")
	asymmetrictDecryptionError     error = errors.New("Asymmetric decryption failed.")
	symmetrictDecryptionError      error = errors.New("Ssymmetric decryption failed.")
	payloadDecodeError             error = errors.New("Payload decoding failed.")
	payloadDecryptionError         error = errors.New("Payload decryption failed.")
	invalidPayloadError            error = errors.New("Invalid payload provided.")
	keyNotFoundError               error = errors.New("Symmetric Key not found by ID.")
	invalidSignatureEncodingError  error = errors.New("Invalid signature encoding.")
	invalidIssuerSignatureError    error = errors.New("Invalid issuer signature provided.")
	invalidCertifierSignatureError error = errors.New("Invalid certifier signature provided.")
)

/*
	Primitives
*/

var rng io.Reader = rand.Reader

func Base64EncodeToString(src []byte) string {
	return base64.StdEncoding.EncodeToString(src)
}

func Base64DecodeString(src string) (res []byte, err error) {
	res, err = base64.StdEncoding.DecodeString(src)
	if err != nil {
		return nil, base64DecodeError
	}
	return
}

func ValidateNonce(nonce []byte) error {
	if len(nonce) != SymmetricNonceSize {
		return invalidNonceError
	}
	return nil
}

func ValidateSymmetricKey(key []byte) error {
	if len(key) != SymmetricKeySize {
		return invalidSymmetricKeyError
	}
	return nil
}

func Hash(plaintext []byte) []byte {
	hashed := sha256.Sum256(plaintext)
	return hashed[:]
}

func Sign(key *rsa.PrivateKey, plaintext []byte) ([]byte, error) {
	signature, err := rsa.SignPKCS1v15(rng, key, HashingAlgorithm, plaintext[:])
	if err != nil {
		return nil, signError
	}
	return signature, nil
}

func Verify(key *rsa.PublicKey, plaintext []byte, signature []byte) bool {
	err := rsa.VerifyPKCS1v15(key, HashingAlgorithm, plaintext[:], signature)
	return err == nil
}

func AsymmetricEncrypt(key *rsa.PublicKey, plaintext []byte) ([]byte, error) {
	ciphertext, err := rsa.EncryptPKCS1v15(rng, key, plaintext)
	if err != nil {
		return nil, asymmetrictEncryptionError
	}
	return ciphertext, nil
}

func AsymmetricDecrypt(key *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	plaintext, err := rsa.DecryptPKCS1v15(rng, key, ciphertext)
	if err != nil {
		return nil, asymmetrictDecryptionError
	}
	return plaintext, nil
}

func NewAead(key []byte) (cipher.AEAD, error) {
	aead, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, aeadCreationError
	}
	return aead, nil
}

func SymmetricEncrypt(aead cipher.AEAD, dst []byte, nonce []byte, plaintext []byte) []byte {
	return aead.Seal(
		dst,
		nonce,
		plaintext,
		[]byte{},
	)
}

func SymmetricDecrypt(aead cipher.AEAD, dst []byte, nonce []byte, ciphertext []byte) ([]byte, error) {
	plaintext, err := aead.Open(
		dst,
		nonce,
		ciphertext,
		[]byte{},
	)
	if err != nil {
		return nil, symmetrictDecryptionError
	}
	return plaintext, nil
}

/*
	Transaction decryption
*/
func (op *Transaction) Decrypt(asymKey *rsa.PrivateKey) (*Operation, error) {
	// Base64 decode payload
	payloadBytes, err := Base64DecodeString(op.Payload)
	if err != nil {
		return nil, payloadDecodeError
	}

	// Decrypt payload if encrypted
	var aead cipher.AEAD = nil
	if op.Encryption.Encrypted {

		// Check nonce
		symKeyNonceBytes, err := Base64DecodeString(op.Encryption.Nonce)
		if err == nil {
			err = ValidateNonce(symKeyNonceBytes)
		}
		if err != nil {
			return nil, invalidNonceError
		}

		// Find a symmetric key that passes challenge
		for symKeyCipher, symKeyChallenge := range op.Encryption.Challenges {
			// Decode symmetric key ciphertext
			symKeyCipherBytes, err := Base64DecodeString(symKeyCipher)
			if err != nil {
				continue
			}

			// Decrypt symmetric key
			symKeyPlainBytes, err := AsymmetricDecrypt(asymKey, symKeyCipherBytes)
			if err == nil {
				err = ValidateSymmetricKey(symKeyPlainBytes)
			}
			if err != nil {
				continue
			}

			// Decode challenge
			symKeyAead, _ := NewAead(symKeyPlainBytes)
			symKeyChallengeBytes, err := Base64DecodeString(symKeyChallenge)
			if err != nil {
				continue
			}

			// Decrypt challenge
			decryptedChallenge, decryptedChallengeErr := SymmetricDecrypt(
				symKeyAead,
				symKeyChallengeBytes[:0],
				symKeyNonceBytes,
				symKeyChallengeBytes,
			)

			// Test if decrypted challenge is correct
			if decryptedChallengeErr == nil &&
				string(decryptedChallenge) == CorrectChallenge {
				aead = symKeyAead
				break
			}
		}

		// No symmetric keys worked
		if aead == nil {
			return nil, noSymmetricKeyFoundError
		}

		// Decrypt payload
		payloadBytes, _ = SymmetricDecrypt(
			aead,
			payloadBytes[:0],
			symKeyNonceBytes,
			payloadBytes,
		)
	}

	// Decode payload into structure
	var decodedOp Operation
	payloadDecodeErr := decodedOp.Decode(payloadBytes)
	if payloadDecodeErr != nil {
		return nil, invalidPayloadError
	}

	return &decodedOp, nil
}

/*
	Permanent decryption
*/
func (op *Operation) Decrypt(
	decrypt Decryptor,
) ([]byte, error) {
	// Base64 decode payload
	payloadBytes, err := Base64DecodeString(op.Payload)
	if err != nil {
		return nil, payloadDecodeError
	}

	// Decrypt payload
	if op.Encryption.Encrypted {
		// Decode nonce
		nonceBytes, err := Base64DecodeString(op.Encryption.Nonce)
		if err == nil {
			err = ValidateNonce(nonceBytes)
		}
		if err != nil {
			return nil, invalidNonceError
		}

		// Decrypt
		payloadBytes, err = decrypt(op.Encryption.KeyId, nonceBytes, payloadBytes)
		if err != nil {
			return nil, keyNotFoundError
		}
	}

	return payloadBytes, nil
}

/*
	Signature verification
*/
func (op *Operation) Verify(
	issuerSigningKey *rsa.PublicKey,
	certifierSigningKey *rsa.PublicKey,
	payload []byte,
) (verified error) {
	verified = decodeAndVerifySignature(issuerSigningKey, op.Issue.Signature, payload, invalidIssuerSignatureError)
	if verified != nil {
		return
	}
	verified = decodeAndVerifySignature(certifierSigningKey, op.Certification.Signature, payload, invalidCertifierSignatureError)
	return
}
func decodeAndVerifySignature(
	signingKey *rsa.PublicKey,
	signatureEncoded string,
	payload []byte,
	invalidSignatureError error,
) error {
	// Decode signature
	var signature []byte
	var err error
	if signature, err = Base64DecodeString(signatureEncoded); err != nil {
		return invalidSignatureEncodingError
	}

	// Verify signature
	if verified := Verify(signingKey, Hash(payload), signature); !verified {
		return invalidSignatureError
	}
	return nil
}

/*
	Defines the set of signers that were verified
*/
type VerifiedSigners struct {
	IssuerId    string
	CertifierId string
}
