/*
	Test helpers
*/

package users

import (
	"encoding/json"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"bytes"
	"testing"
	"time"
)

/*
	General
*/

func booleanToString(boolean bool) string {
	if boolean {
		return "true"
	}
	return "false"
}


/*
	Crypto
*/

func generatePrivateKey() *rsa.PrivateKey {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	return priv
}

func generatePublicKey() *rsa.PublicKey {
	priv := generatePrivateKey()
	return &priv.PublicKey
}

func pemEncodeKey(key *rsa.PublicKey) string {
	keyBytes, _ := x509.MarshalPKIXPublicKey(key)
	block := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: keyBytes,
	}
	buf := new(bytes.Buffer)
	pem.Encode(buf, block)
	return string(pem.EncodeToMemory(block))
}

func jsonPemEncodeKey(key *rsa.PublicKey) string {
	res, _ := json.Marshal(pemEncodeKey(key))
	return string(res)
}

/*
	Server
*/
func multipleWorkersConfig() Config {
	return Config{
		NumWorkers: 6,
	}
}

func singleWorkerConfig() Config {
	return Config{
		NumWorkers: 1,
	}
}

func resetAndStartServer(t *testing.T, conf Config) bool {
	serverSingleton = server{}
	err := StartServer(conf)
	if err != nil {
		t.Errorf(err.Error())
		return false
	}
	return true
}

/*
	Creation requests
*/
func generateUserCreateRequest(
	issuerId string,
	certifierId string,
	userId string,
	channelAddPermission bool,
	userAddPermission bool,
	userRemovePermission bool,
	userEncKeyUpdatePermission bool,
	userSignKeyUpdatePermission bool,
	userPermissionsUpdatePermission bool,
) (request []byte, object *UserObject) {
	// Encode keys
	encKey := generatePublicKey()
	encKeyStringEncoded := jsonPemEncodeKey(encKey)
	var encKeyStringDecoded string
	json.Unmarshal([]byte(encKeyStringEncoded), &encKeyStringDecoded)
	signKey := generatePublicKey()
	signKeyStringEncoded := jsonPemEncodeKey(signKey)
	var signKeyStringDecoded string
	json.Unmarshal([]byte(signKeyStringEncoded), &signKeyStringDecoded)

	// Permissions strings
	channelAddPermissionString := booleanToString(channelAddPermission)
	userAddPermissionString := booleanToString(userAddPermission)
	userRemovePermissionString := booleanToString(userRemovePermission)
	userEncKeyUpdatePermissionString := booleanToString(userEncKeyUpdatePermission)
	userSignKeyUpdatePermissionString := booleanToString(userSignKeyUpdatePermission)
	userPermissionsUpdatePermissionString := booleanToString(userPermissionsUpdatePermission)

	request = []byte(`{
		"type": 0,
		"issuerId": "` + issuerId + `",
		"certifierId": "` + certifierId + `",
		"fields": [],
		"data": {
			"id": "` + userId + `",
			"encKey": ` + encKeyStringEncoded + `,
			"signKey": ` + signKeyStringEncoded + `,
			"permissions": {
				"channel": {
					"add": ` + channelAddPermissionString + `
				},
				"user": {
					"add": ` + userAddPermissionString + `,
					"remove": ` + userRemovePermissionString + `,
					"encKeyUpdate": ` + userEncKeyUpdatePermissionString + `,
					"signKeyUpdate": ` + userSignKeyUpdatePermissionString + `,
					"permissionsUpdate": ` + userPermissionsUpdatePermissionString + `
				}
			},
			"active": true
		},
		"timestamp": "2018-01-01T00:00:00Z"
	}`)

	object = &UserObject{
		Id:      userId,
		EncKey:  encKeyStringDecoded,
		SignKey: signKeyStringDecoded,
		Permissions: PermissionsObject{
			Channel: ChannelPermissionsObject{
				Add: channelAddPermission,
			},
			User: UserPermissionsObject{
				Add:               userAddPermission,
				Remove:            userRemovePermission,
				EncKeyUpdate:      userEncKeyUpdatePermission,
				SignKeyUpdate:     userSignKeyUpdatePermission,
				PermissionsUpdate: userPermissionsUpdatePermission,
			},
		},
		Active:     true,
		CreatedAt:  time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC),
		DisabledAt: time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC),
	}

	return
}

func makeUserCreationRequest(
	skipVerification bool,
	issuerId string,
	certifierId string,
	userId string,
	channelAddPermission bool,
	userAddPermission bool,
	userRemovePermission bool,
	userEncKeyUpdatePermission bool,
	userSignKeyUpdatePermission bool,
	userPermissionsUpdatePermission bool,
) (chan *UserResponse, *UserObject, []error) {
	// Generate user payload and expected object
	requestBytes, userObjectPtr := generateUserCreateRequest(
		issuerId,
		certifierId,
		userId,
		channelAddPermission,
		userAddPermission,
		userRemovePermission,
		userEncKeyUpdatePermission,
		userSignKeyUpdatePermission,
		userPermissionsUpdatePermission,
	)

	// Make a request to the server
	requestFunc := MakeRequest
	if skipVerification {
		requestFunc = MakeUnverifiedRequest
	}
	channel, errs := requestFunc(requestBytes)
	return channel, userObjectPtr, errs
}

func makeAndGetUserCreationRequest(
	t *testing.T,
	skipVerification bool,
	issuerId string,
	certifierId string,
	userId string,
	channelAddPermission bool,
	userAddPermission bool,
	userRemovePermission bool,
	userEncKeyUpdatePermission bool,
	userSignKeyUpdatePermission bool,
	userPermissionsUpdatePermission bool,
) (*UserResponse, bool, *UserObject, bool) {
	// Generate user payload and expected object
	channel, userObjectPtr, errs := makeUserCreationRequest(
		skipVerification,
		issuerId,
		certifierId,
		userId,
		channelAddPermission,
		userAddPermission,
		userRemovePermission,
		userEncKeyUpdatePermission,
		userSignKeyUpdatePermission,
		userPermissionsUpdatePermission,
	)
	if len(errs) != 0 {
		t.Errorf("Valid create request should go through, errs=%v", errs)
		return nil, false, nil, false
	}

	serverResponsePtr, ok := <-channel
	return serverResponsePtr, ok, userObjectPtr, true
}

func createUser(
	t *testing.T,
	skipVerification bool,
	issuerId string,
	certifierId string,
	userId string,
	channelAddPermission bool,
	userAddPermission bool,
	userRemovePermission bool,
	userEncKeyUpdatePermission bool,
	userSignKeyUpdatePermission bool,
	userPermissionsUpdatePermission bool,
) (*UserObject, bool) {
	serverResponsePtr, ok, userObjectPtr, success := makeAndGetUserCreationRequest(
		t,
		skipVerification,
		issuerId,
		certifierId,
		userId,
		channelAddPermission,
		userAddPermission,
		userRemovePermission,
		userEncKeyUpdatePermission,
		userSignKeyUpdatePermission,
		userPermissionsUpdatePermission,
	)
	if !success {
		return nil, false
	}
	if !ok || serverResponsePtr.Result != Success {
		t.Errorf("Valid create request should succeed, result:%v", *serverResponsePtr)
		return nil, false
	}
	return userObjectPtr, true
}

func createUnverifiedUser(
	t *testing.T,
	userId string,
	channelAddPermission bool,
	userAddPermission bool,
	userRemovePermission bool,
	userEncKeyUpdatePermission bool,
	userSignKeyUpdatePermission bool,
	userPermissionsUpdatePermission bool,
) bool {
	_, success := createUser(
		t, true, "NONE", "NONE", userId,
		channelAddPermission,
		userAddPermission,
		userRemovePermission,
		userEncKeyUpdatePermission,
		userSignKeyUpdatePermission,
		userPermissionsUpdatePermission,
	)
	if !success {
		return false
	}
	return true
}

func createIssuerAndCertifier(
	t *testing.T,

	issuerChannelAddPermission bool,
	issuerUserAddPermission bool,
	issuerUserRemovePermission bool,
	issuerUserEncKeyUpdatePermission bool,
	issuerUserSignKeyUpdatePermission bool,
	issuerUserPermissionsUpdatePermission bool,

	certifierChannelAddPermission bool,
	certifierUserAddPermission bool,
	certifierUserRemovePermission bool,
	certifierUserEncKeyUpdatePermission bool,
	certifierUserSignKeyUpdatePermission bool,
	certifierUserPermissionsUpdatePermission bool,
) bool {
	if !createUnverifiedUser(
		t,
		"ISSUER",
		issuerChannelAddPermission,
		issuerUserAddPermission,
		issuerUserRemovePermission,
		issuerUserEncKeyUpdatePermission,
		issuerUserSignKeyUpdatePermission,
		issuerUserPermissionsUpdatePermission,
	) {
		return false
	}
	if !createUnverifiedUser(
		t,
		"CERTIFIER",
		certifierChannelAddPermission,
		certifierUserAddPermission,
		certifierUserRemovePermission,
		certifierUserEncKeyUpdatePermission,
		certifierUserSignKeyUpdatePermission,
		certifierUserPermissionsUpdatePermission,
	) {
		return false
	}
	return true
}

/*
	Read requests
*/

func generateUserReadRequest(issuerId string, certifierId string, users []string) (request []byte) {
	var usersJsonString string
	usersJson, _ := json.Marshal(users)
	usersJsonString = string(usersJson)
	return []byte(`{
		"type": 2,
		"issuerId": "` + issuerId + `",
		"certifierId": "` + certifierId + `",
		"fields": ` + usersJsonString + `
	}`)
}

func makeUserReadRequest(issuerId string, certifierId string, users []string) (chan *UserResponse, []error) {
	requestBytes := generateUserReadRequest(issuerId, certifierId, users)
	return MakeRequest(requestBytes)
}

func makeAndGetUserReadRequest(t *testing.T, issuerId string, certifierId string, users []string) (*UserResponse, bool, bool) {
	channel, errs := makeUserReadRequest(issuerId, certifierId, users)
	if len(errs) > 0 {
		t.Errorf("Valid read request should go through\n. errs=%v", errs)
		return nil, false, false
	}
	serverResponsePtr, ok := <-channel
	return serverResponsePtr, ok, true
}
