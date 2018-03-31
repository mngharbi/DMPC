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
		recipientKey = generatePrivateKey()
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

func GeneratePermanentEncryptedOperation(
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

func GeneratePermanentEncryptedOperationWithEncryption(
	keyId string,
	permanentKey []byte,
	permanentNonce []byte,
	requestType int,
	plainPayload []byte,
	modifyIssuerSignature func([]byte) ([]byte, bool),
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
	issuerKey := generatePrivateKey()
	certifierKey := generatePrivateKey()
	issuerSignature, _ := Sign(issuerKey, plainPayloadHashed[:])
	issuerSignature, issuerSignatureEncoded := modifyIssuerSignature(issuerSignature)

	certifierSignature, _ := Sign(certifierKey, plainPayloadHashed[:])
	certifierSignature, certifierSignatureEncoded := modifyCertifierSignature(certifierSignature)

	return GeneratePermanentEncryptedOperation(
		true,
		keyId,
		permanentNonce,
		false,
		issuerSignature,
		issuerSignatureEncoded,
		certifierSignature,
		certifierSignatureEncoded,
		requestType,
		ciphertextPayload,
		false,
	), issuerKey, certifierKey
}

const invalidBase64string = "12"
const validBase64string = "bQ=="

func dummyByteToByteTransformer(str []byte) ([]byte, bool) {
	return str, false
}

/*
	Temporary decryption
*/

func TestTemporaryValidOperation(t *testing.T) {
	// Make valid encrypted operation
	encryptedInnerOperation, _, _ := GeneratePermanentEncryptedOperationWithEncryption(
		"KEY_ID",
		generateRandomBytes(SymmetricKeySize),
		generateRandomBytes(SymmetricNonceSize),
		1,
		[]byte("REQUEST_PAYLOAD"),
		dummyByteToByteTransformer,
		dummyByteToByteTransformer,
	)
	innerOperationJson, _ := encryptedInnerOperation.Encode()
	temporaryEncryptedOperation, recipientKey := GenerateTemporaryEncryptedOperationWithEncryption(
		innerOperationJson,
		[]byte(correctChallenge),
		func(challenges map[string]string) {
			challenges[validBase64string] = validBase64string
		},
		nil,
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
	temporaryEncryptedOperation := GenerateTemporaryEncryptedOperation(
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
	temporaryEncryptedOperation := GenerateTemporaryEncryptedOperation(
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
	encryptedInnerOperation, _, _ := GeneratePermanentEncryptedOperationWithEncryption(
		"KEY_ID",
		generateRandomBytes(SymmetricKeySize),
		generateRandomBytes(SymmetricNonceSize),
		1,
		[]byte("REQUEST_PAYLOAD"),
		dummyByteToByteTransformer,
		dummyByteToByteTransformer,
	)
	innerOperationJson, _ := encryptedInnerOperation.Encode()

	// Use invalid base64 string for nonce
	temporaryEncryptedOperation := GenerateTemporaryEncryptedOperation(
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
	temporaryEncryptedOperation = GenerateTemporaryEncryptedOperation(
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

func TestTemporaryInavlidChallenges(t *testing.T) {
	// Make valid encrypted operation
	encryptedInnerOperation, _, _ := GeneratePermanentEncryptedOperationWithEncryption(
		"KEY_ID",
		generateRandomBytes(SymmetricKeySize),
		generateRandomBytes(SymmetricNonceSize),
		1,
		[]byte("REQUEST_PAYLOAD"),
		dummyByteToByteTransformer,
		dummyByteToByteTransformer,
	)
	innerOperationJson, _ := encryptedInnerOperation.Encode()

	privateKey := generatePrivateKey()

	// Invalid symmetric key encoding
	challenges := map[string]string{
		invalidBase64string: invalidBase64string,
	}
	temporaryEncryptedOperation := GenerateTemporaryEncryptedOperation(
		true,
		challenges,
		generateRandomBytes(SymmetricNonceSize),
		false,
		innerOperationJson,
		false,
	)
	_, err := temporaryEncryptedOperation.Decrypt(privateKey)
	if err != noSymmetricKeyFoundError {
		t.Errorf("Temporary decryption should fail with invalid temp key encoding. err=%v", err)
		return
	}

	// Invalid symmetric key ciphertext
	invalidKeyCiphertext := generateRandomBytes(1 + maxAsymmetricCiphertextLength)
	invalidKeyCiphertextBase64 := Base64EncodeToString(invalidKeyCiphertext)
	challenges = map[string]string{
		invalidKeyCiphertextBase64: validBase64string,
	}
	temporaryEncryptedOperation = GenerateTemporaryEncryptedOperation(
		true,
		challenges,
		generateRandomBytes(SymmetricNonceSize),
		false,
		innerOperationJson,
		false,
	)
	_, err = temporaryEncryptedOperation.Decrypt(privateKey)
	if err != noSymmetricKeyFoundError {
		t.Errorf("Temporary decryption should fail with invalid temp key ciphertext. err=%v", err)
		return
	}

	// Invalid symmetric key length
	invalidKeyCiphertext = generateRandomBytes(1 + SymmetricKeySize)
	invalidKeyCiphertextBase64 = Base64EncodeToString(invalidKeyCiphertext)
	challenges = map[string]string{
		invalidKeyCiphertextBase64: validBase64string,
	}
	temporaryEncryptedOperation = GenerateTemporaryEncryptedOperation(
		true,
		challenges,
		generateRandomBytes(SymmetricNonceSize),
		false,
		innerOperationJson,
		false,
	)
	_, err = temporaryEncryptedOperation.Decrypt(privateKey)
	if err != noSymmetricKeyFoundError {
		t.Errorf("Temporary decryption should fail with invalid temp key ciphertext. err=%v", err)
		return
	}

	// Invalid symmetric key length
	invalidKeyCiphertext = generateRandomBytes(1 + SymmetricKeySize)
	invalidKeyCiphertextBase64 = Base64EncodeToString(invalidKeyCiphertext)
	challenges = map[string]string{
		invalidKeyCiphertextBase64: validBase64string,
	}
	temporaryEncryptedOperation = GenerateTemporaryEncryptedOperation(
		true,
		challenges,
		generateRandomBytes(SymmetricNonceSize),
		false,
		innerOperationJson,
		false,
	)
	_, err = temporaryEncryptedOperation.Decrypt(privateKey)
	if err != noSymmetricKeyFoundError {
		t.Errorf("Temporary decryption should fail with invalid temp key ciphertext. err=%v", err)
		return
	}

	// Invalid challenge ciphertext encoding
	validKeyCiphertext := generateRandomBytes(SymmetricKeySize)
	validKeyCiphertextBase64 := Base64EncodeToString(validKeyCiphertext)
	challenges = map[string]string{
		validKeyCiphertextBase64: invalidBase64string,
	}
	temporaryEncryptedOperation = GenerateTemporaryEncryptedOperation(
		true,
		challenges,
		generateRandomBytes(SymmetricNonceSize),
		false,
		innerOperationJson,
		false,
	)
	_, err = temporaryEncryptedOperation.Decrypt(privateKey)
	if err != noSymmetricKeyFoundError {
		t.Errorf("Temporary decryption should fail with invalid challege encoding. err=%v", err)
		return
	}

	// Invalid challenge ciphertext
	validKeyCiphertext = generateRandomBytes(SymmetricKeySize)
	validKeyCiphertextBase64 = Base64EncodeToString(validKeyCiphertext)
	invalidChallengeCiphertext := generateRandomBytes(1 + SymmetricKeySize)
	invalidChallengeCiphertextBase64 := Base64EncodeToString(invalidChallengeCiphertext)
	challenges = map[string]string{
		validKeyCiphertextBase64: invalidChallengeCiphertextBase64,
	}
	temporaryEncryptedOperation = GenerateTemporaryEncryptedOperation(
		true,
		challenges,
		generateRandomBytes(SymmetricNonceSize),
		false,
		innerOperationJson,
		false,
	)
	_, err = temporaryEncryptedOperation.Decrypt(privateKey)
	if err != noSymmetricKeyFoundError {
		t.Errorf("Temporary decryption should fail with invalid challenge ciphertext. err=%v", err)
		return
	}

	// Valid challenge ciphertext, doesn't match challenge string
	temporaryEncryptedOperation, _ = GenerateTemporaryEncryptedOperationWithEncryption(
		innerOperationJson,
		[]byte("WRONG CHALLENGE"),
		func(map[string]string) {},
		privateKey,
	)
	_, err = temporaryEncryptedOperation.Decrypt(privateKey)
	if err != noSymmetricKeyFoundError {
		t.Errorf("Temporary decryption should fail with incorrect challenge. err=%v", err)
		return
	}

	// Skipping wrong challenge
	temporaryEncryptedOperation, _ = GenerateTemporaryEncryptedOperationWithEncryption(
		innerOperationJson,
		[]byte(correctChallenge),
		func(challenges map[string]string) {
			challenges[invalidBase64string] = invalidBase64string
		},
		privateKey,
	)
	_, err = temporaryEncryptedOperation.Decrypt(privateKey)
	if err != nil {
		t.Errorf("Temporary decryption should not fail with extra incorrect challenge. err=%v", err)
		return
	}
}

func TestTemporaryInavlidPayloadStruncture(t *testing.T) {
	// Make undecryptable permanent operation
	temporaryEncryptedOperation, privateKey := GenerateTemporaryEncryptedOperationWithEncryption(
		[]byte("{"),
		[]byte(correctChallenge),
		func(map[string]string) {},
		nil,
	)

	_, err := temporaryEncryptedOperation.Decrypt(privateKey)
	if err != invalidPayloadError {
		t.Errorf("Temporary decryption should fail with incorrectly structured payload. err=%v", err)
	}
}

/*
	Permanent decryption
*/
func TestPermanentValidOperation(t *testing.T) {
	// Make valid encrypted operation
	permanentKey := generateRandomBytes(SymmetricKeySize)
	permanentNonce := generateRandomBytes(SymmetricNonceSize)
	requestPayload := []byte("REQUEST_PAYLOAD")
	encryptedInnerOperation, issuerKey, certifierKey := GeneratePermanentEncryptedOperationWithEncryption(
		"KEY_ID",
		permanentKey,
		permanentNonce,
		1,
		requestPayload,
		dummyByteToByteTransformer,
		dummyByteToByteTransformer,
	)

	decryptedDecodedPayload, err := encryptedInnerOperation.Decrypt(
		func(string) []byte { return permanentKey },
		&issuerKey.PublicKey,
		&certifierKey.PublicKey,
	)
	if err != nil ||
		!reflect.DeepEqual(decryptedDecodedPayload, requestPayload) {
		t.Errorf("Permanent decryption failed. found=%+v, expected=%v", decryptedDecodedPayload, requestPayload)
		return
	}
}

func TestPermanentInvalidPayload(t *testing.T) {
	// Make valid encrypted operation
	encryptedInnerOperation := GeneratePermanentEncryptedOperation(
		true,
		"KEY_ID",
		generateRandomBytes(SymmetricNonceSize),
		true,
		[]byte(validBase64string),
		true,
		[]byte(validBase64string),
		true,
		1,
		[]byte(invalidBase64string),
		true,
	)

	_, err := encryptedInnerOperation.Decrypt(
		func(string) []byte { return generateRandomBytes(SymmetricKeySize) },
		nil,
		nil,
	)
	if err != payloadDecodeError {
		t.Errorf("Permanent decryption should fail with invalid base64 payload. found=%+v, expected=%v")
		return
	}
}

func TestPermanentInvalidNonce(t *testing.T) {
	// Make permanent opration with invalid nonce encoding
	encryptedInnerOperation := GeneratePermanentEncryptedOperation(
		true,
		"KEY_ID",
		[]byte(invalidBase64string),
		true,
		[]byte(validBase64string),
		true,
		[]byte(validBase64string),
		true,
		1,
		[]byte(validBase64string),
		true,
	)

	_, err := encryptedInnerOperation.Decrypt(
		func(string) []byte { return generateRandomBytes(SymmetricKeySize) },
		nil,
		nil,
	)
	if err != invalidNonceError {
		t.Errorf("Permanent decryption should fail with invalid base64 nonce.")
		return
	}

	// Make valid encrypted operation
	encryptedInnerOperation = GeneratePermanentEncryptedOperation(
		true,
		"KEY_ID",
		generateRandomBytes(1+SymmetricNonceSize),
		false,
		[]byte(validBase64string),
		true,
		[]byte(validBase64string),
		true,
		1,
		[]byte(validBase64string),
		true,
	)

	_, err = encryptedInnerOperation.Decrypt(
		func(string) []byte { return generateRandomBytes(SymmetricKeySize) },
		nil,
		nil,
	)
	if err != invalidNonceError {
		t.Errorf("Permanent decryption should fail with invalid nonce length.")
		return
	}
}

func TestPermanentNotFoundKey(t *testing.T) {
	// Make valid permanent opration
	encryptedInnerOperation := GeneratePermanentEncryptedOperation(
		true,
		"KEY_ID",
		generateRandomBytes(SymmetricNonceSize),
		false,
		[]byte(validBase64string),
		true,
		[]byte(validBase64string),
		true,
		1,
		[]byte(validBase64string),
		true,
	)

	// Decrypt with function that returns no key
	_, err := encryptedInnerOperation.Decrypt(
		func(string) []byte { return nil },
		nil,
		nil,
	)
	if err != keyNotFoundError {
		t.Errorf("Permanent decryption should fail with no key found.")
		return
	}
}

func TestPermanentInvalidIssuerSignature(t *testing.T) {
	// Make permanent opration with invalid issuer signature
	permanentKey := generateRandomBytes(SymmetricKeySize)
	permanentNonce := generateRandomBytes(SymmetricNonceSize)
	requestPayload := []byte("REQUEST_PAYLOAD")
	encryptedInnerOperation, issuerKey, certifierKey := GeneratePermanentEncryptedOperationWithEncryption(
		"KEY_ID",
		permanentKey,
		permanentNonce,
		1,
		requestPayload,
		func([]byte) ([]byte, bool) { return []byte(invalidBase64string), true },
		dummyByteToByteTransformer,
	)

	_, err := encryptedInnerOperation.Decrypt(
		func(string) []byte { return permanentKey },
		&issuerKey.PublicKey,
		&certifierKey.PublicKey,
	)
	if err != invalidSignatureEncodingError {
		t.Errorf("Permanent decryption should fail with invalid base64 issuer signature. err=%v", err)
	}

	// Make permanent opration without corresponding issuer signature
	encryptedInnerOperation, issuerKey, certifierKey = GeneratePermanentEncryptedOperationWithEncryption(
		"KEY_ID",
		permanentKey,
		permanentNonce,
		1,
		requestPayload,
		func([]byte) ([]byte, bool) { return []byte(validBase64string), true },
		dummyByteToByteTransformer,
	)

	_, err = encryptedInnerOperation.Decrypt(
		func(string) []byte { return permanentKey },
		&issuerKey.PublicKey,
		&certifierKey.PublicKey,
	)
	if err != invalidIssuerSignatureError {
		t.Errorf("Permanent decryption should fail with invalid issuer signature.")
	}
}

func TestPermanentInvalidCertifierSignature(t *testing.T) {
	// Make permanent opration with invalid certifier signature
	permanentKey := generateRandomBytes(SymmetricKeySize)
	permanentNonce := generateRandomBytes(SymmetricNonceSize)
	requestPayload := []byte("REQUEST_PAYLOAD")
	encryptedInnerOperation, issuerKey, certifierKey := GeneratePermanentEncryptedOperationWithEncryption(
		"KEY_ID",
		permanentKey,
		permanentNonce,
		1,
		requestPayload,
		dummyByteToByteTransformer,
		func([]byte) ([]byte, bool) { return []byte(invalidBase64string), true },
	)

	_, err := encryptedInnerOperation.Decrypt(
		func(string) []byte { return permanentKey },
		&issuerKey.PublicKey,
		&certifierKey.PublicKey,
	)
	if err != invalidSignatureEncodingError {
		t.Errorf("Permanent decryption should fail with invalid base64 certifier signature. err=%v", err)
	}

	// Make permanent opration without corresponding certifier signature
	encryptedInnerOperation, issuerKey, certifierKey = GeneratePermanentEncryptedOperationWithEncryption(
		"KEY_ID",
		permanentKey,
		permanentNonce,
		1,
		requestPayload,
		dummyByteToByteTransformer,
		func([]byte) ([]byte, bool) { return []byte(validBase64string), true },
	)

	_, err = encryptedInnerOperation.Decrypt(
		func(string) []byte { return permanentKey },
		&issuerKey.PublicKey,
		&certifierKey.PublicKey,
	)
	if err != invalidCertifierSignatureError {
		t.Errorf("Permanent decryption should fail with invalid certifier signature.")
	}
}
