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
	encryptedOperation, _, _ := GenerateOperationWithEncryption(
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
	innerOperationJson, _ := encryptedOperation.Encode()
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
		!reflect.DeepEqual(encryptedOperation, decryptedTransaction) {
		t.Errorf("Transaction decryption failed.")
		t.Errorf("encryptedOperation=%v", encryptedOperation)
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
	encryptedOperation, _, _ := GenerateOperationWithEncryption(
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
	innerOperationJson, _ := encryptedOperation.Encode()

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
	encryptedOperation, _, _ := GenerateOperationWithEncryption(
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
	innerOperationJson, _ := encryptedOperation.Encode()

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
	// Make undecryptable operation
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
	encryptedOperation, _, _ := GenerateOperationWithEncryption(
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

	decryptedDecodedPayload, err := encryptedOperation.Decrypt(
		DecryptorFunctor(map[string][]byte{"KEY_ID": permanentKey}, true),
	)
	if err != nil ||
		!reflect.DeepEqual(decryptedDecodedPayload, requestPayload) {
		t.Errorf("Permanent decryption failed. found=%+v, expected=%v", decryptedDecodedPayload, requestPayload)
		return
	}
}

func TestPermanentInvalidPayload(t *testing.T) {
	// Make valid encrypted operation
	encryptedOperation := GenerateOperation(
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

	_, err := encryptedOperation.Decrypt(
		DecryptorFunctor(map[string][]byte{"KEY_ID": generateRandomBytes(SymmetricKeySize)}, true),
	)
	if err != payloadDecodeError {
		t.Errorf("Permanent decryption should fail with invalid base64 payload.")
		return
	}
}

func TestPermanentInvalidNonce(t *testing.T) {
	// Make operation with invalid nonce encoding
	encryptedOperation := GenerateOperation(
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

	_, err := encryptedOperation.Decrypt(
		DecryptorFunctor(map[string][]byte{"KEY_ID": generateRandomBytes(SymmetricKeySize)}, true),
	)
	if err != invalidNonceError {
		t.Errorf("Permanent decryption should fail with invalid base64 nonce.")
		return
	}

	// Make valid encrypted operation
	encryptedOperation = GenerateOperation(
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

	_, err = encryptedOperation.Decrypt(
		DecryptorFunctor(map[string][]byte{"KEY_ID": generateRandomBytes(SymmetricKeySize)}, true),
	)
	if err != invalidNonceError {
		t.Errorf("Permanent decryption should fail with invalid nonce length.")
		return
	}
}

func TestPermanentNotFoundKey(t *testing.T) {
	// Make valid operation
	encryptedOperation := GenerateOperation(
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
	_, err := encryptedOperation.Decrypt(
		DecryptorFunctor(nil, false),
	)
	if err != keyNotFoundError {
		t.Errorf("Permanent decryption should fail with no key found.")
		return
	}
}

func TestPermanentInvalidIssuerSignature(t *testing.T) {
	// Make operation with invalid issuer signature
	permanentKey := generateRandomBytes(SymmetricKeySize)
	permanentNonce := generateRandomBytes(SymmetricNonceSize)
	requestPayload := []byte("REQUEST_PAYLOAD")
	encryptedOperation, issuerKey, certifierKey := GenerateOperationWithEncryption(
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

	payload, err := encryptedOperation.Decrypt(
		DecryptorFunctor(map[string][]byte{"KEY_ID": permanentKey}, true),
	)
	if err != nil {
		t.Errorf("Permanent decryption should not fail with invalid base64 issuer signature. err=%v", err)
	}
	err = encryptedOperation.Verify(
		&issuerKey.PublicKey,
		&certifierKey.PublicKey,
		payload,
	)
	if err != invalidSignatureEncodingError {
		t.Errorf("Verify should fail with invalid base64 issuer signature. err=%v", err)
	}

	// Make operation without corresponding issuer signature
	encryptedOperation, issuerKey, certifierKey = GenerateOperationWithEncryption(
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

	payload, err = encryptedOperation.Decrypt(
		DecryptorFunctor(map[string][]byte{"KEY_ID": permanentKey}, true),
	)
	if err != nil {
		t.Errorf("Permanent decryption should not fail with invalid issuer signature.")
	}
	err = encryptedOperation.Verify(
		&issuerKey.PublicKey,
		&certifierKey.PublicKey,
		payload,
	)
	if err != invalidIssuerSignatureError {
		t.Errorf("Permanent decryption should fail with invalid issuer signature.")
	}
}

func TestPermanentInvalidCertifierSignature(t *testing.T) {
	// Make operation with invalid certifier signature
	permanentKey := generateRandomBytes(SymmetricKeySize)
	permanentNonce := generateRandomBytes(SymmetricNonceSize)
	requestPayload := []byte("REQUEST_PAYLOAD")
	encryptedOperation, issuerKey, certifierKey := GenerateOperationWithEncryption(
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

	payload, err := encryptedOperation.Decrypt(
		DecryptorFunctor(map[string][]byte{"KEY_ID": permanentKey}, true),
	)
	if err != nil {
		t.Errorf("Permanent decryption should not fail with invalid base64 certifier signature. err=%v", err)
	}
	err = encryptedOperation.Verify(
		&issuerKey.PublicKey,
		&certifierKey.PublicKey,
		payload,
	)
	if err != invalidSignatureEncodingError {
		t.Errorf("Verify should fail with invalid base64 certifier signature. err=%v", err)
	}

	// Make operation without corresponding certifier signature
	encryptedOperation, issuerKey, certifierKey = GenerateOperationWithEncryption(
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

	payload, err = encryptedOperation.Decrypt(
		DecryptorFunctor(map[string][]byte{"KEY_ID": permanentKey}, true),
	)
	if err != nil {
		t.Errorf("Permanent decryption should  notfail with invalid certifier signature.")
	}
	err = encryptedOperation.Verify(
		&issuerKey.PublicKey,
		&certifierKey.PublicKey,
		payload,
	)
	if err != invalidCertifierSignatureError {
		t.Errorf("Verify should fail with invalid base64 certifier signature. err=%v", err)
	}
}
