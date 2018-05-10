package core

import (
	"reflect"
	"testing"
)

/*
	Test helpers
*/

const invalidBase64string = "12"
const validBase64string = "bQ=="

func dummyByteToByteTransformer(str []byte) ([]byte, bool) {
	return str, false
}

/*
	Transaction decryption
*/

func TestValidTransaction(t *testing.T) {
	// Make valid encrypted operation
	encryptedInnerOperation, _, _ := GeneratePermanentEncryptedOperationWithEncryption(
		"KEY_ID",
		generateRandomBytes(SymmetricKeySize),
		generateRandomBytes(SymmetricNonceSize),
		1,
		[]byte("REQUEST_PAYLOAD"),
		"ISSUER",
		dummyByteToByteTransformer,
		"CERTIFIER",
		dummyByteToByteTransformer,
	)
	innerOperationJson, _ := encryptedInnerOperation.Encode()
	transaction, recipientKey := GenerateTransactionWithEncryption(
		innerOperationJson,
		[]byte(CorrectChallenge),
		func(challenges map[string]string) {
			challenges[validBase64string] = validBase64string
		},
		nil,
	)

	decryptedTransaction, err := transaction.Decrypt(recipientKey)
	if err != nil ||
		!reflect.DeepEqual(encryptedInnerOperation, decryptedTransaction) {
		t.Errorf("Transaction decryption failed.")
		t.Errorf("encryptedInnerOperation=%v", encryptedInnerOperation)
		t.Errorf("decryptedTransaction=%v", decryptedTransaction)
		t.Errorf("err=%v", err)
	}
}

func TestInavlidTransactionPayloadEncoding(t *testing.T) {
	// Use invalid base64 string for payload
	transaction := GenerateTransaction(
		false,
		map[string]string{},
		[]byte("PLAINTEXT"),
		false,
		[]byte(invalidBase64string),
		true,
	)

	_, err := transaction.Decrypt(GeneratePrivateKey())
	if err != payloadDecodeError {
		t.Errorf("Transaction decryption should fail with invalid payload encoding. err=%v", err)
	}
}

func TestInavlidTransactionPayloadStructure(t *testing.T) {
	// Use invalid payload structure
	transaction := GenerateTransaction(
		false,
		map[string]string{},
		[]byte("PLAINTEXT"),
		false,
		[]byte("INVALID_PAYLOAD"),
		false,
	)

	_, err := transaction.Decrypt(GeneratePrivateKey())
	if err != invalidPayloadError {
		t.Errorf("Transaction decryption should fail with invalid payload structure. err=%v", err)
	}
}

func TestInavlidTransactionNonce(t *testing.T) {
	// Make valid encrypted operation
	encryptedInnerOperation, _, _ := GeneratePermanentEncryptedOperationWithEncryption(
		"KEY_ID",
		generateRandomBytes(SymmetricKeySize),
		generateRandomBytes(SymmetricNonceSize),
		1,
		[]byte("REQUEST_PAYLOAD"),
		"ISSUER",
		dummyByteToByteTransformer,
		"CERTIFIER",
		dummyByteToByteTransformer,
	)
	innerOperationJson, _ := encryptedInnerOperation.Encode()

	// Use invalid base64 string for nonce
	transaction := GenerateTransaction(
		true,
		map[string]string{},
		[]byte(invalidBase64string),
		true,
		innerOperationJson,
		false,
	)

	_, err := transaction.Decrypt(GeneratePrivateKey())
	if err != invalidNonceError {
		t.Errorf("Transaction decryption should fail with invalid nonce encoding. err=%v", err)
		return
	}

	// Use nonce with invalid length
	transaction = GenerateTransaction(
		true,
		map[string]string{},
		generateRandomBytes(1+SymmetricNonceSize),
		false,
		innerOperationJson,
		false,
	)

	_, err = transaction.Decrypt(GeneratePrivateKey())
	if err != invalidNonceError {
		t.Errorf("Transaction decryption should fail with invalid nonce length. err=%v", err)
		return
	}
}

func TestInavlidTransactionChallenges(t *testing.T) {
	// Make valid encrypted operation
	encryptedInnerOperation, _, _ := GeneratePermanentEncryptedOperationWithEncryption(
		"KEY_ID",
		generateRandomBytes(SymmetricKeySize),
		generateRandomBytes(SymmetricNonceSize),
		1,
		[]byte("REQUEST_PAYLOAD"),
		"ISSUER",
		dummyByteToByteTransformer,
		"CERTIFIER",
		dummyByteToByteTransformer,
	)
	innerOperationJson, _ := encryptedInnerOperation.Encode()

	privateKey := GeneratePrivateKey()

	// Invalid symmetric key encoding
	challenges := map[string]string{
		invalidBase64string: invalidBase64string,
	}
	transaction := GenerateTransaction(
		true,
		challenges,
		generateRandomBytes(SymmetricNonceSize),
		false,
		innerOperationJson,
		false,
	)
	_, err := transaction.Decrypt(privateKey)
	if err != noSymmetricKeyFoundError {
		t.Errorf("Transaction decryption should fail with invalid temp key encoding. err=%v", err)
		return
	}

	// Invalid symmetric key ciphertext
	invalidKeyCiphertext := generateRandomBytes(1 + maxAsymmetricCiphertextLength)
	invalidKeyCiphertextBase64 := Base64EncodeToString(invalidKeyCiphertext)
	challenges = map[string]string{
		invalidKeyCiphertextBase64: validBase64string,
	}
	transaction = GenerateTransaction(
		true,
		challenges,
		generateRandomBytes(SymmetricNonceSize),
		false,
		innerOperationJson,
		false,
	)
	_, err = transaction.Decrypt(privateKey)
	if err != noSymmetricKeyFoundError {
		t.Errorf("Transaction decryption should fail with invalid temp key ciphertext. err=%v", err)
		return
	}

	// Invalid symmetric key length
	invalidKeyCiphertext = generateRandomBytes(1 + SymmetricKeySize)
	invalidKeyCiphertextBase64 = Base64EncodeToString(invalidKeyCiphertext)
	challenges = map[string]string{
		invalidKeyCiphertextBase64: validBase64string,
	}
	transaction = GenerateTransaction(
		true,
		challenges,
		generateRandomBytes(SymmetricNonceSize),
		false,
		innerOperationJson,
		false,
	)
	_, err = transaction.Decrypt(privateKey)
	if err != noSymmetricKeyFoundError {
		t.Errorf("Transaction decryption should fail with invalid temp key ciphertext. err=%v", err)
		return
	}

	// Invalid symmetric key length
	invalidKeyCiphertext = generateRandomBytes(1 + SymmetricKeySize)
	invalidKeyCiphertextBase64 = Base64EncodeToString(invalidKeyCiphertext)
	challenges = map[string]string{
		invalidKeyCiphertextBase64: validBase64string,
	}
	transaction = GenerateTransaction(
		true,
		challenges,
		generateRandomBytes(SymmetricNonceSize),
		false,
		innerOperationJson,
		false,
	)
	_, err = transaction.Decrypt(privateKey)
	if err != noSymmetricKeyFoundError {
		t.Errorf("Transaction decryption should fail with invalid temp key ciphertext. err=%v", err)
		return
	}

	// Invalid challenge ciphertext encoding
	validKeyCiphertext := generateRandomBytes(SymmetricKeySize)
	validKeyCiphertextBase64 := Base64EncodeToString(validKeyCiphertext)
	challenges = map[string]string{
		validKeyCiphertextBase64: invalidBase64string,
	}
	transaction = GenerateTransaction(
		true,
		challenges,
		generateRandomBytes(SymmetricNonceSize),
		false,
		innerOperationJson,
		false,
	)
	_, err = transaction.Decrypt(privateKey)
	if err != noSymmetricKeyFoundError {
		t.Errorf("Transaction decryption should fail with invalid challege encoding. err=%v", err)
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
	transaction = GenerateTransaction(
		true,
		challenges,
		generateRandomBytes(SymmetricNonceSize),
		false,
		innerOperationJson,
		false,
	)
	_, err = transaction.Decrypt(privateKey)
	if err != noSymmetricKeyFoundError {
		t.Errorf("Transaction decryption should fail with invalid challenge ciphertext. err=%v", err)
		return
	}

	// Valid challenge ciphertext, doesn't match challenge string
	transaction, _ = GenerateTransactionWithEncryption(
		innerOperationJson,
		[]byte("WRONG CHALLENGE"),
		func(map[string]string) {},
		privateKey,
	)
	_, err = transaction.Decrypt(privateKey)
	if err != noSymmetricKeyFoundError {
		t.Errorf("Transaction decryption should fail with incorrect challenge. err=%v", err)
		return
	}

	// Skipping wrong challenge
	transaction, _ = GenerateTransactionWithEncryption(
		innerOperationJson,
		[]byte(CorrectChallenge),
		func(challenges map[string]string) {
			challenges[invalidBase64string] = invalidBase64string
		},
		privateKey,
	)
	_, err = transaction.Decrypt(privateKey)
	if err != nil {
		t.Errorf("Transaction decryption should not fail with extra incorrect challenge. err=%v", err)
		return
	}
}

func TestInavlidTransactionPayloadStruncture(t *testing.T) {
	// Make undecryptable permanent operation
	transaction, privateKey := GenerateTransactionWithEncryption(
		[]byte("{"),
		[]byte(CorrectChallenge),
		func(map[string]string) {},
		nil,
	)

	_, err := transaction.Decrypt(privateKey)
	if err != invalidPayloadError {
		t.Errorf("Transaction decryption should fail with incorrectly structured payload. err=%v", err)
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
	encryptedInnerOperation, _, _ := GeneratePermanentEncryptedOperationWithEncryption(
		"KEY_ID",
		permanentKey,
		permanentNonce,
		1,
		requestPayload,
		"ISSUER",
		dummyByteToByteTransformer,
		"CERTIFIER",
		dummyByteToByteTransformer,
	)

	decryptedDecodedPayload, err := encryptedInnerOperation.Decrypt(
		func(string) []byte { return permanentKey },
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
		"ISSUER",
		[]byte(validBase64string),
		true,
		"CERTIFIER",
		[]byte(validBase64string),
		true,
		1,
		[]byte(invalidBase64string),
		true,
	)

	_, err := encryptedInnerOperation.Decrypt(
		func(string) []byte { return generateRandomBytes(SymmetricKeySize) },
	)
	if err != payloadDecodeError {
		t.Errorf("Permanent decryption should fail with invalid base64 payload.")
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
		"ISSUER",
		[]byte(validBase64string),
		true,
		"CERTIFIER",
		[]byte(validBase64string),
		true,
		1,
		[]byte(validBase64string),
		true,
	)

	_, err := encryptedInnerOperation.Decrypt(
		func(string) []byte { return generateRandomBytes(SymmetricKeySize) },
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
		"ISSUER",
		[]byte(validBase64string),
		true,
		"CERTIFIER",
		[]byte(validBase64string),
		true,
		1,
		[]byte(validBase64string),
		true,
	)

	_, err = encryptedInnerOperation.Decrypt(
		func(string) []byte { return generateRandomBytes(SymmetricKeySize) },
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
		"ISSUER",
		[]byte(validBase64string),
		true,
		"CERTIFIER",
		[]byte(validBase64string),
		true,
		1,
		[]byte(validBase64string),
		true,
	)

	// Decrypt with function that returns no key
	_, err := encryptedInnerOperation.Decrypt(
		func(string) []byte { return nil },
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
		"ISSUER",
		func([]byte) ([]byte, bool) { return []byte(invalidBase64string), true },
		"CERTIFIER",
		dummyByteToByteTransformer,
	)

	payload, err := encryptedInnerOperation.Decrypt(
		func(string) []byte { return permanentKey },
	)
	if err != nil {
		t.Errorf("Permanent decryption should not fail with invalid base64 issuer signature. err=%v", err)
	}
	err = encryptedInnerOperation.Verify(
		&issuerKey.PublicKey,
		&certifierKey.PublicKey,
		payload,
	)
	if err != invalidSignatureEncodingError {
		t.Errorf("Verify should fail with invalid base64 issuer signature. err=%v", err)
	}

	// Make permanent opration without corresponding issuer signature
	encryptedInnerOperation, issuerKey, certifierKey = GeneratePermanentEncryptedOperationWithEncryption(
		"KEY_ID",
		permanentKey,
		permanentNonce,
		1,
		requestPayload,
		"ISSUER",
		func([]byte) ([]byte, bool) { return []byte(validBase64string), true },
		"CERTIFIER",
		dummyByteToByteTransformer,
	)

	payload, err = encryptedInnerOperation.Decrypt(
		func(string) []byte { return permanentKey },
	)
	if err != nil {
		t.Errorf("Permanent decryption should not fail with invalid issuer signature.")
	}
	err = encryptedInnerOperation.Verify(
		&issuerKey.PublicKey,
		&certifierKey.PublicKey,
		payload,
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
		"ISSUER",
		dummyByteToByteTransformer,
		"CERTIFIER",
		func([]byte) ([]byte, bool) { return []byte(invalidBase64string), true },
	)

	payload, err := encryptedInnerOperation.Decrypt(
		func(string) []byte { return permanentKey },
	)
	if err != nil {
		t.Errorf("Permanent decryption should not fail with invalid base64 certifier signature. err=%v", err)
	}
	err = encryptedInnerOperation.Verify(
		&issuerKey.PublicKey,
		&certifierKey.PublicKey,
		payload,
	)
	if err != invalidSignatureEncodingError {
		t.Errorf("Verify should fail with invalid base64 certifier signature. err=%v", err)
	}

	// Make permanent opration without corresponding certifier signature
	encryptedInnerOperation, issuerKey, certifierKey = GeneratePermanentEncryptedOperationWithEncryption(
		"KEY_ID",
		permanentKey,
		permanentNonce,
		1,
		requestPayload,
		"ISSUER",
		dummyByteToByteTransformer,
		"CERTIFIER",
		func([]byte) ([]byte, bool) { return []byte(validBase64string), true },
	)

	payload, err = encryptedInnerOperation.Decrypt(
		func(string) []byte { return permanentKey },
	)
	if err != nil {
		t.Errorf("Permanent decryption should  notfail with invalid certifier signature.")
	}
	err = encryptedInnerOperation.Verify(
		&issuerKey.PublicKey,
		&certifierKey.PublicKey,
		payload,
	)
	if err != invalidCertifierSignatureError {
		t.Errorf("Verify should fail with invalid base64 certifier signature. err=%v", err)
	}
}
