/*
	Cryptography utilities
*/

package core

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
)

func generateRandomBytes(nbBytes int) (bytes []byte) {
	bytes = make([]byte, nbBytes)
	rand.Read(bytes)
	return
}


/*
	Key encoding
*/
func pemEncodeBlock(keyBytes []byte, typeString string) string {
	// Build pem block containing key
	block := &pem.Block{
		Type:  typeString,
		Bytes: keyBytes,
	}

	// PEM encode block
	buf := new(bytes.Buffer)
	pem.Encode(buf, block)

	// Return string representing bytes
	return string(pem.EncodeToMemory(block))
}

func PublicAsymKeyToString(key *rsa.PublicKey) string {
	// Break into bytes
	keyBytes, _ := x509.MarshalPKIXPublicKey(key)

	// Encode block
	return pemEncodeBlock(keyBytes, "RSA PUBLIC KEY")
}

func PublicStringToAsymKey(rsaString string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(rsaString))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, errors.New("failed to parse DER encoded public key: " + err.Error())
	}

	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		return nil, errors.New("unknown type of public key" + err.Error())
	}
}

func PrivateAsymKeyToString(key *rsa.PrivateKey) string {
	// Break into bytes
	keyBytes := x509.MarshalPKCS1PrivateKey(key)

	// Encode block
	return pemEncodeBlock(keyBytes, "RSA PRIVATE KEY")
}

func PrivateStringToAsymKey(rsaString string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(rsaString))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the public key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, errors.New("failed to parse PKCS1 encoded private key: " + err.Error())
	}
	return priv, nil
}

/*
	Key generation
*/
func GeneratePrivateKey() *rsa.PrivateKey {
	priv, _ := rsa.GenerateKey(rand.Reader, AsymmetricKeySizeBits)
	return priv
}

func GeneratePublicKey() *rsa.PublicKey {
	priv := GeneratePrivateKey()
	return &priv.PublicKey
}

func GenerateTemporaryEncryptedOperation(
	encrypted bool,
	challenges map[string]string,
	nonce []byte,
	nonceEncoded bool,
	payload []byte,
	payloadEncoded bool,
) *TemporaryEncryptedOperation {
	nonceResult := string(nonce)
	payloadResult := string(payload)
	if !nonceEncoded {
		nonceResult = Base64EncodeToString(nonce)
	}
	if !payloadEncoded {
		payloadResult = Base64EncodeToString(payload)
	}

	return &TemporaryEncryptedOperation{
		Version: 0.1,
		Encryption: TemporaryEncryptionFields{
			Encrypted:  encrypted,
			Challenges: challenges,
			Nonce:      nonceResult,
		},
		Payload: payloadResult,
	}
}

func GenerateTemporaryEncryptedOperationWithEncryption(
	plainPayload []byte,
	plaintextChallenge []byte,
	modifyChallenges func(map[string]string),
	recipientKey *rsa.PrivateKey,
) (*TemporaryEncryptedOperation, *rsa.PrivateKey) {
	// Make temporary key and nonce
	temporaryNonce := generateRandomBytes(SymmetricNonceSize)
	temporaryKey := generateRandomBytes(SymmetricKeySize)

	// Encrypt challenge string and payload using temporary symmetric key
	aead, _ := NewAead(temporaryKey)
	payloadCiphertext := SymmetricEncrypt(
		aead,
		[]byte{},
		temporaryNonce,
		plainPayload,
	)
	challengeCiphertext := SymmetricEncrypt(
		aead,
		[]byte{},
		temporaryNonce,
		[]byte(plaintextChallenge),
	)

	// Make RSA key if nil and use it to encrypt temporary key
	if recipientKey == nil {
		recipientKey = GeneratePrivateKey()
	}
	symKeyEncrypted, _ := AsymmetricEncrypt(&recipientKey.PublicKey, temporaryKey[:])

	// Make challenges map
	challengeCiphertextBase64 := Base64EncodeToString(challengeCiphertext)
	symKeyEncryptedBase64 := Base64EncodeToString(symKeyEncrypted)
	challenges := map[string]string{
		symKeyEncryptedBase64: challengeCiphertextBase64,
	}
	modifyChallenges(challenges)

	return GenerateTemporaryEncryptedOperation(
		true,
		challenges,
		temporaryNonce,
		false,
		payloadCiphertext,
		false,
	), recipientKey
}

/*
	Encrypted Operation(s) generation
*/
func GeneratePermanentEncryptedOperation(
	encrypted bool,
	keyId string,
	nonce []byte,
	nonceEncoded bool,
	issuerId string,
	issuerSignature []byte,
	issuerSignatureEncoded bool,
	certifierId string,
	certifierSignature []byte,
	certifierSignatureEncoded bool,
	requestType int,
	payload []byte,
	payloadEncoded bool,
) *PermanentEncryptedOperation {
	// Encode or convert to string
	nonceResult := string(nonce)
	issuerSignatureResult := string(issuerSignature)
	certifierSignatureResult := string(certifierSignature)
	payloadResult := string(payload)
	if !nonceEncoded {
		nonceResult = Base64EncodeToString(nonce)
	}
	if !issuerSignatureEncoded {
		issuerSignatureResult = Base64EncodeToString(issuerSignature)
	}
	if !certifierSignatureEncoded {
		certifierSignatureResult = Base64EncodeToString(certifierSignature)
	}
	if !payloadEncoded {
		payloadResult = Base64EncodeToString(payload)
	}

	// Create operation
	return &PermanentEncryptedOperation{
		Encryption: PermanentEncryptionFields{
			Encrypted: encrypted,
			KeyId:     keyId,
			Nonce:     nonceResult,
		},
		Issue: PermanentAuthenticationFields{
			Id:        issuerId,
			Signature: issuerSignatureResult,
		},
		Certification: PermanentAuthenticationFields{
			Id:        certifierId,
			Signature: certifierSignatureResult,
		},
		Meta: PermanentMetaFields{
			RequestType: requestType,
		},
		Payload: payloadResult,
	}
}

func GeneratePermanentEncryptedOperationWithEncryption(
	keyId string,
	permanentKey []byte,
	permanentNonce []byte,
	requestType int,
	plainPayload []byte,
	issuerId string,
	modifyIssuerSignature func([]byte) ([]byte, bool),
	certifierId string,
	modifyCertifierSignature func([]byte) ([]byte, bool),
) (*PermanentEncryptedOperation, *rsa.PrivateKey, *rsa.PrivateKey) {
	// Encrypt payload with symmetric permanent key
	aead, _ := NewAead(permanentKey)
	ciphertextPayload := SymmetricEncrypt(
		aead,
		[]byte{},
		permanentNonce,
		plainPayload,
	)

	// Hash and sign plaintext payload with new RSA keys
	plainPayloadHashed := Hash(plainPayload)
	issuerKey := GeneratePrivateKey()
	certifierKey := GeneratePrivateKey()
	issuerSignature, _ := Sign(issuerKey, plainPayloadHashed[:])
	issuerSignature, issuerSignatureEncoded := modifyIssuerSignature(issuerSignature)

	certifierSignature, _ := Sign(certifierKey, plainPayloadHashed[:])
	certifierSignature, certifierSignatureEncoded := modifyCertifierSignature(certifierSignature)

	return GeneratePermanentEncryptedOperation(
		true,
		keyId,
		permanentNonce,
		false,
		issuerId,
		issuerSignature,
		issuerSignatureEncoded,
		certifierId,
		certifierSignature,
		certifierSignatureEncoded,
		requestType,
		ciphertextPayload,
		false,
	), issuerKey, certifierKey
}
