package core

import (
	"crypto/rand"
	"crypto/rsa"
	"reflect"
	"testing"
)

/*
	Test helpers
*/

func generatePrivateKey() *rsa.PrivateKey {
	priv, _ := rsa.GenerateKey(rand.Reader, AsymmetricKeySizeBits)
	return priv
}

func generateRandomBytes(nbBytes int) (bytes []byte) {
	bytes = make([]byte, nbBytes)
	rand.Read(bytes)
	return
}

func generateTemporaryEncryptedOperation(
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

func generateTemporaryEncryptedOperationWithEncryption(
	plainPayload []byte,
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
		[]byte(correctChallenge),
	)

	// Make RSA key and use it to encrypt temporary key
	recipientKey := generatePrivateKey()
	symKeyEncrypted, _ := AsymmetricEncrypt(&recipientKey.PublicKey, temporaryKey[:])

	// Make challenges map
	challengeCiphertextBase64 := Base64EncodeToString(challengeCiphertext)
	symKeyEncryptedBase64 := Base64EncodeToString(symKeyEncrypted)
	challenges := map[string]string{
		"random":              "test",
		symKeyEncryptedBase64: challengeCiphertextBase64,
		"random2":             "test2",
	}

	return generateTemporaryEncryptedOperation(
		true,
		challenges,
		temporaryNonce,
		false,
		payloadCiphertext,
		false,
	), recipientKey
}

func generatePermanentEncryptedOperation(
	encrypted bool,
	keyId string,
	nonce []byte,
	nonceEncoded bool,
	issuerSignature []byte,
	issuerSignatureEncoded bool,
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
			Signature: issuerSignatureResult,
		},
		Certification: PermanentAuthenticationFields{
			Signature: certifierSignatureResult,
		},
		Meta: PermanentMetaFields{
			RequestType: requestType,
		},
		Payload: payloadResult,
	}
}

func generatePermanentEncryptedOperationWithEncryption(
	keyId string,
	permanentKey []byte,
	requestType int,
	plainPayload []byte,
) *PermanentEncryptedOperation {
	// Encrypt payload with symmetric permanent key
	permanentNonce := generateRandomBytes(SymmetricNonceSize)
	aead, _ := NewAead(permanentKey)
	ciphertextPayload := SymmetricEncrypt(
		aead,
		[]byte{},
		permanentNonce,
		plainPayload,
	)

	// Hash and sign plaintext payload with new RSA keys
	plainPayloadHashed := Hash(plainPayload)
	issuerKey := generatePrivateKey()
	certifierKey := generatePrivateKey()
	issuerSignature, _ := Sign(issuerKey, plainPayloadHashed[:])
	certifierSignature, _ := Sign(certifierKey, plainPayloadHashed[:])

	return generatePermanentEncryptedOperation(
		true,
		keyId,
		permanentNonce,
		false,
		issuerSignature,
		false,
		certifierSignature,
		false,
		requestType,
		ciphertextPayload,
		false,
	)
}

const invalidBase64string = "12"

/*
	Temporary decryption
*/

func TestTemporaryValidOperation(t *testing.T) {
	// Make valid encrypted operation
	encryptedInnerOperation := generatePermanentEncryptedOperationWithEncryption(
		"KEY_ID",
		generateRandomBytes(SymmetricKeySize),
		1,
		[]byte("REQUEST_PAYLOAD"),
	)
	innerOperationJson, _ := encryptedInnerOperation.Encode()
	temporaryEncryptedOperation, recipientKey := generateTemporaryEncryptedOperationWithEncryption(
		innerOperationJson,
	)

	decryptedTemporaryEncryptedOperation, err := temporaryEncryptedOperation.Decrypt(recipientKey)
	if err != nil ||
		!reflect.DeepEqual(encryptedInnerOperation, decryptedTemporaryEncryptedOperation) {
		t.Errorf("Temporary decryption failed.")
		t.Errorf("encryptedInnerOperation=%v", encryptedInnerOperation)
		t.Errorf("decryptedTemporaryEncryptedOperation=%v", decryptedTemporaryEncryptedOperation)
		t.Errorf("err=%v", err)
	}
}

func TestTemporaryInavlidPayloadEncoding(t *testing.T) {
	// Use invalid base64 string for payload
	temporaryEncryptedOperation := generateTemporaryEncryptedOperation(
		false,
		map[string]string{},
		[]byte("PLAINTEXT"),
		false,
		[]byte(invalidBase64string),
		true,
	)

	_, err := temporaryEncryptedOperation.Decrypt(generatePrivateKey())
	if err != payloadDecodeError {
		t.Errorf("Temporary decryption should fail with invalid payload encoding. err=%v", err)
	}
}

func TestTemporaryInavlidPayloadStructure(t *testing.T) {
	// Use invalid payload structure
	temporaryEncryptedOperation := generateTemporaryEncryptedOperation(
		false,
		map[string]string{},
		[]byte("PLAINTEXT"),
		false,
		[]byte("INVALID_PAYLOAD"),
		false,
	)

	_, err := temporaryEncryptedOperation.Decrypt(generatePrivateKey())
	if err != invalidPayloadError {
		t.Errorf("Temporary decryption should fail with invalid payload structure. err=%v", err)
	}
}

func TestTemporaryInavlidNonce(t *testing.T) {
	// Make valid encrypted operation
	encryptedInnerOperation := generatePermanentEncryptedOperationWithEncryption(
		"KEY_ID",
		generateRandomBytes(SymmetricKeySize),
		1,
		[]byte("REQUEST_PAYLOAD"),
	)
	innerOperationJson, _ := encryptedInnerOperation.Encode()

	// Use invalid base64 string for nonce
	temporaryEncryptedOperation := generateTemporaryEncryptedOperation(
		true,
		map[string]string{},
		[]byte(invalidBase64string),
		true,
		innerOperationJson,
		false,
	)

	_, err := temporaryEncryptedOperation.Decrypt(generatePrivateKey())
	if err != invalidNonceError {
		t.Errorf("Temporary decryption should fail with invalid nonce encoding. err=%v", err)
		return
	}

	// Use nonce with invalid length
	temporaryEncryptedOperation = generateTemporaryEncryptedOperation(
		true,
		map[string]string{},
		generateRandomBytes(1+SymmetricNonceSize),
		false,
		innerOperationJson,
		false,
	)

	_, err = temporaryEncryptedOperation.Decrypt(generatePrivateKey())
	if err != invalidNonceError {
		t.Errorf("Temporary decryption should fail with invalid nonce length. err=%v", err)
		return
	}
}
