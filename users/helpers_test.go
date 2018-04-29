/*
	Test helpers
*/

package users

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"github.com/mngharbi/DMPC/core"
	"strings"
	"testing"
	"time"
)

/*
	General
*/

func getJanuaryDate(date int) time.Time {
	return time.Date(2018, time.January, date, 0, 0, 0, 0, time.UTC)
}

func booleanToString(boolean bool) string {
	if boolean {
		return "true"
	}
	return "false"
}

/*
	Crypto
*/

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
	err := StartServer(conf, log, shutdownProgram)
	if err != nil {
		t.Errorf(err.Error())
		return false
	}
	return true
}

// @TODO: Change request helpers to use vararg syntax
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
	encKey := core.GeneratePublicKey()
	encKeyStringEncoded := jsonPemEncodeKey(encKey)
	var encKeyStringDecoded string
	json.Unmarshal([]byte(encKeyStringEncoded), &encKeyStringDecoded)
	signKey := core.GeneratePublicKey()
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

	// Default date
	defaultDate := getJanuaryDate(15)

	request = []byte(`{
		"type": 0,
		"issuerId": "` + issuerId + `",
		"certifierId": "` + certifierId + `",
		"fields": [],
		` + generateJsonForTimePtr("timestamp", &defaultDate) + `
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
		}
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
		CreatedAt:  defaultDate,
		DisabledAt: defaultDate,
		UpdatedAt:  defaultDate,
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
	Create requests
*/

func generateJsonForStringPtr(id string, strPtr *string) (res string) {
	res = ""
	if strPtr != nil {
		res = `"` + id + `": "` + *strPtr + `",`
	}
	return
}

func generateJsonForBoolPtr(id string, boolPtr *bool) (res string) {
	res = ""
	if boolPtr != nil {
		res = `"` + id + `": ` + booleanToString(*boolPtr) + `,`
	}
	return
}
func generateJsonForTimePtr(id string, timePtr *time.Time) (res string) {
	res = ""
	if timePtr != nil {
		res = `"` + id + `": "` + timePtr.Format(time.RFC3339) + `",`
	}
	return
}

func generateUserUpdateRequest(
	issuerId string,
	certifierId string,
	fields []string,
	timestamp time.Time,

	idPtr *string,
	encKeyPtr *string,
	signKeyPtr *string,
	channelAddPermissionPtr *bool,
	userAddPermissionPtr *bool,
	userRemovePermissionPtr *bool,
	userEncKeyUpdatePermissionPtr *bool,
	userSignKeyUpdatePermissionPtr *bool,
	userPermissionsUpdatePtr *bool,
	activePtr *bool,
	createdAtPtr *time.Time,
	disabledAtPtr *time.Time,
	updatedAtPtr *time.Time,
) (request []byte) {
	fieldsJson, _ := json.Marshal(fields)
	fieldsJsonString := string(fieldsJson)

	// Make minimal json equivalent for base level fields
	baseLevelJson := ""
	baseLevelStringPtrs := map[string]*string{
		"id":      idPtr,
		"encKey":  encKeyPtr,
		"signKey": signKeyPtr,
	}
	for id, stringPtr := range baseLevelStringPtrs {
		baseLevelJson += generateJsonForStringPtr(id, stringPtr)
	}
	baseLevelJson += generateJsonForBoolPtr("active", activePtr)
	baseLevelTimePtrs := map[string]*time.Time{
		"createdAt":  createdAtPtr,
		"disabledAt": disabledAtPtr,
		"updatedAt":  updatedAtPtr,
	}
	for id, timePtr := range baseLevelTimePtrs {
		baseLevelJson += generateJsonForTimePtr(id, timePtr)
	}

	// Make minimal json equivalent for user permissions
	userPermissionsStr := ""
	userPermissionsBoolPtrs := map[string]*bool{
		"add":               userAddPermissionPtr,
		"remove":            userRemovePermissionPtr,
		"encKeyUpdate":      userEncKeyUpdatePermissionPtr,
		"signKeyUpdate":     userSignKeyUpdatePermissionPtr,
		"permissionsUpdate": userPermissionsUpdatePtr,
	}
	for id, boolPtr := range userPermissionsBoolPtrs {
		userPermissionsStr += generateJsonForBoolPtr(id, boolPtr)
	}
	if len(userPermissionsStr) > 0 {
		userPermissionsStr = strings.TrimSuffix(userPermissionsStr, ",")
		userPermissionsStr = `"user": {` + userPermissionsStr + `},`
	}

	// Make minimal json equivalent for channel permissions
	channelAddPermissionStr := generateJsonForBoolPtr("add", channelAddPermissionPtr)
	if channelAddPermissionPtr != nil {
		channelAddPermissionStr = strings.TrimSuffix(channelAddPermissionStr, ",")
		channelAddPermissionStr = `"channel": {` + channelAddPermissionStr + `},`
	}

	// Combine permissions json
	permissionsStr := userPermissionsStr + channelAddPermissionStr
	if len(permissionsStr) > 0 {
		permissionsStr = strings.TrimSuffix(permissionsStr, ",")
		permissionsStr = `"permissions": {` + permissionsStr + `},`
	}

	// Combine everything together
	dataStr := baseLevelJson + permissionsStr
	if len(dataStr) > 0 {
		dataStr = strings.TrimSuffix(dataStr, ",")
		dataStr = `"data": {` + dataStr + `}`
	}

	return []byte(`{
		"type": 1,
		` + generateJsonForTimePtr("timestamp", &timestamp) + `
		"issuerId": "` + issuerId + `",
		"certifierId": "` + certifierId + `",
		"fields": ` + fieldsJsonString + `,
		` + dataStr + `
	}`)
}

func makeUserUpdateRequest(
	issuerId string,
	certifierId string,
	fields []string,
	timestamp time.Time,

	idPtr *string,
	encKeyPtr *string,
	signKeyPtr *string,
	channelAddPermissionPtr *bool,
	userAddPermissionPtr *bool,
	userRemovePermissionPtr *bool,
	userEncKeyUpdatePermissionPtr *bool,
	userSignKeyUpdatePermissionPtr *bool,
	userPermissionsUpdatePtr *bool,
	activePtr *bool,
	createdAtPtr *time.Time,
	disabledAtPtr *time.Time,
	updatedAtPtr *time.Time,
) (chan *UserResponse, []error) {
	requestBytes := generateUserUpdateRequest(
		issuerId, certifierId, fields, timestamp,
		idPtr, encKeyPtr, signKeyPtr,
		channelAddPermissionPtr, userAddPermissionPtr, userRemovePermissionPtr,
		userEncKeyUpdatePermissionPtr, userSignKeyUpdatePermissionPtr, userPermissionsUpdatePtr,
		activePtr, createdAtPtr, disabledAtPtr, updatedAtPtr,
	)
	return MakeRequest(requestBytes)
}

func makeAndGetUserUpdateRequest(
	t *testing.T,

	issuerId string,
	certifierId string,
	fields []string,
	timestamp time.Time,

	idPtr *string,
	encKeyPtr *string,
	signKeyPtr *string,
	channelAddPermissionPtr *bool,
	userAddPermissionPtr *bool,
	userRemovePermissionPtr *bool,
	userEncKeyUpdatePermissionPtr *bool,
	userSignKeyUpdatePermissionPtr *bool,
	userPermissionsUpdatePtr *bool,
	activePtr *bool,
	createdAtPtr *time.Time,
	disabledAtPtr *time.Time,
	updatedAtPtr *time.Time,
) (*UserResponse, bool, bool) {
	channel, errs := makeUserUpdateRequest(
		issuerId, certifierId, fields, timestamp,
		idPtr, encKeyPtr, signKeyPtr,
		channelAddPermissionPtr, userAddPermissionPtr, userRemovePermissionPtr,
		userEncKeyUpdatePermissionPtr, userSignKeyUpdatePermissionPtr, userPermissionsUpdatePtr,
		activePtr, createdAtPtr, disabledAtPtr, updatedAtPtr,
	)
	if len(errs) > 0 {
		t.Errorf("Valid update request should go through\n. errs=%v", errs)
		return nil, false, false
	}
	serverResponsePtr, ok := <-channel
	return serverResponsePtr, ok, true
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
