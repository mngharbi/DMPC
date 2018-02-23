package users

import (
	"crypto/rsa"
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

/*
	Helpers
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

func generateUserCreateRequest(issuerId string, certifierId string, userId string, userAddPermission bool) (request []byte, object *UserObject, encKey *rsa.PublicKey, signKey *rsa.PublicKey) {
	encKey = generatePublicKey()
	encKeyStringEncoded := jsonPemEncodeKey(encKey)
	var encKeyStringDecoded string
	json.Unmarshal([]byte(encKeyStringEncoded), &encKeyStringDecoded)
	signKey = generatePublicKey()
	signKeyStringEncoded := jsonPemEncodeKey(signKey)
	var signKeyStringDecoded string
	json.Unmarshal([]byte(signKeyStringEncoded), &signKeyStringDecoded)
	userAddPermissionString := "false"
	if userAddPermission {
		userAddPermissionString = "true"
	}

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
					"add": true
				},
				"user": {
					"add": ` + userAddPermissionString + `,
					"remove": false,
					"encKeyUpdate": false,
					"signKeyUpdate": false,
					"permissionsUpdate": false
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
				Add: true,
			},
			User: UserPermissionsObject{
				Add:               userAddPermission,
				Remove:            false,
				EncKeyUpdate:      false,
				SignKeyUpdate:     false,
				PermissionsUpdate: false,
			},
		},
		Active:     true,
		CreatedAt:  time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC),
		DisabledAt: time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:  time.Date(2018, time.January, 1, 0, 0, 0, 0, time.UTC),
	}

	return
}

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

func resetAndStartServer(conf Config) error {
	serverSingleton = server{}
	return StartServer(conf)
}

/*
	General tests
*/

func TestStartShutdownSingleWorker(t *testing.T) {
	initErr := resetAndStartServer(singleWorkerConfig())
	if initErr != nil {
		t.Errorf(initErr.Error())
		return
	}
	ShutdownServer()
}

func TestStartShutdown(t *testing.T) {
	initErr := resetAndStartServer(multipleWorkersConfig())
	if initErr != nil {
		t.Errorf(initErr.Error())
		return
	}
	ShutdownServer()
}

func TestMalformattedRequest(t *testing.T) {
	initErr := resetAndStartServer(multipleWorkersConfig())
	if initErr != nil {
		t.Errorf(initErr.Error())
		return
	}

	requestBytes := []byte(`{invalid}`)

	_, errs := MakeRequest(requestBytes)
	if len(errs) == 0 {
		t.Error("Malformatted request should fail")
	}

	ShutdownServer()
}

/*
	Read requests
*/

func TestEmptyReadRequest(t *testing.T) {
	initErr := resetAndStartServer(multipleWorkersConfig())
	if initErr != nil {
		t.Errorf(initErr.Error())
		return
	}

	requestBytes := generateUserReadRequest("ISSUER", "CERTIFIER", []string{})

	_, errs := MakeRequest(requestBytes)
	if len(errs) == 0 {
		t.Error("Read request with no users should fail")
	}

	ShutdownServer()
}

func TestUnknownIssuerReadRequest(t *testing.T) {
	initErr := resetAndStartServer(multipleWorkersConfig())
	if initErr != nil {
		t.Errorf(initErr.Error())
		return
	}

	requestBytes := generateUserReadRequest("ISSUER", "CERTIFIER", []string{"USER"})

	channel, errs := MakeRequest(requestBytes)
	if len(errs) > 0 {
		t.Errorf("Read request with inexistent issuer should go through, errs=%v", errs)
		return
	}
	serverResponsePtr, ok := <-channel
	if !ok || serverResponsePtr.Result != IssuerUnknownError {
		t.Errorf("Read request with inexistent issuer should fail, result:%v", *serverResponsePtr)
		return
	}

	ShutdownServer()
}

func TestUnknownCertifierReadRequest(t *testing.T) {
	initErr := resetAndStartServer(multipleWorkersConfig())
	if initErr != nil {
		t.Errorf(initErr.Error())
		return
	}

	requestBytes, _, _, _ := generateUserCreateRequest("ISSUER", "ISSUER", "ISSUER", true)
	channel, errs := MakeUnverifiedRequest(requestBytes)
	if len(errs) != 0 {
		t.Errorf("Unverified create request with inexistent issuer/certifier should go through, errs=%v", errs)
		return
	}
	serverResponsePtr, ok := <-channel
	if !ok || serverResponsePtr.Result != Success {
		t.Errorf("Unverified create request with inexistent issuer/certifier should succeed, result:%v", *serverResponsePtr)
		return
	}

	requestBytes = generateUserReadRequest("ISSUER", "CERTIFIER", []string{"USER"})
	channel, errs = MakeRequest(requestBytes)
	if len(errs) > 0 {
		t.Errorf("Read request with inexistent certifier should go through, errs=%v", errs)
		return
	}
	serverResponsePtr, ok = <-channel
	if !ok || serverResponsePtr.Result != CertifierUnknownError {
		t.Errorf("Read request with inexistent certifier should fail, result:%v", *serverResponsePtr)
		return
	}

	ShutdownServer()
}

func TestUnknownSubjectReadRequest(t *testing.T) {
	initErr := resetAndStartServer(multipleWorkersConfig())
	if initErr != nil {
		t.Errorf(initErr.Error())
		return
	}

	requestBytes, _, _, _ := generateUserCreateRequest("ISSUER", "ISSUER", "ISSUER", true)
	channel, errs := MakeUnverifiedRequest(requestBytes)
	if len(errs) != 0 {
		t.Errorf("Unverified create request with inexistent issuer/certifier should go through, errs=%v", errs)
		return
	}
	serverResponsePtr, ok := <-channel
	if !ok || serverResponsePtr.Result != Success {
		t.Errorf("Unverified create request with inexistent issuer/certifier should succeed, result:%v", *serverResponsePtr)
		return
	}

	requestBytes, _, _, _ = generateUserCreateRequest("CERTIFIER", "CERTIFIER", "CERTIFIER", true)
	channel, errs = MakeUnverifiedRequest(requestBytes)
	if len(errs) != 0 {
		t.Errorf("Unverified create request with inexistent issuer/certifier should go through, errs=%v", errs)
		return
	}
	serverResponsePtr, ok = <-channel
	if !ok || serverResponsePtr.Result != Success {
		t.Errorf("Unverified create request with inexistent issuer/certifier should succeed, result:%v", *serverResponsePtr)
		return
	}

	requestBytes = generateUserReadRequest("ISSUER", "CERTIFIER", []string{"USER"})
	channel, errs = MakeRequest(requestBytes)
	if len(errs) > 0 {
		t.Errorf("Read request with inexistent certifier should go through, errs=%v", errs)
		return
	}
	serverResponsePtr, ok = <-channel
	if !ok || serverResponsePtr.Result != SubjectUnknownError {
		t.Errorf("Read request with inexistent subject should fail, result:%v", *serverResponsePtr)
		return
	}

	ShutdownServer()
}

func TestExistentUserReadRequest(t *testing.T) {
	initErr := resetAndStartServer(multipleWorkersConfig())
	if initErr != nil {
		t.Errorf(initErr.Error())
		return
	}

	requestBytes, _, _, _ := generateUserCreateRequest("ISSUER", "ISSUER", "ISSUER", false)
	channel, errs := MakeUnverifiedRequest(requestBytes)
	if len(errs) != 0 {
		t.Errorf("Unverified create request with inexistent issuer/certifier should go through, errs=%v", errs)
		return
	}
	serverResponsePtr, ok := <-channel
	if !ok || serverResponsePtr.Result != Success {
		t.Errorf("Unverified create request with inexistent issuer/certifier should succeed, result:%v", *serverResponsePtr)
		return
	}

	requestBytes, _, _, _ = generateUserCreateRequest("CERTIFIER", "CERTIFIER", "CERTIFIER", true)
	channel, errs = MakeUnverifiedRequest(requestBytes)
	if len(errs) != 0 {
		t.Errorf("Unverified create request with inexistent issuer/certifier should go through, errs=%v", errs)
		return
	}
	serverResponsePtr, ok = <-channel
	if !ok || serverResponsePtr.Result != Success {
		t.Errorf("Unverified create request with inexistent issuer/certifier should succeed, result:%v", *serverResponsePtr)
		return
	}

	requestBytes, userObjectPtr, _, _ := generateUserCreateRequest("ISSUER", "CERTIFIER", "USER", true)
	channel, errs = MakeRequest(requestBytes)
	if len(errs) != 0 {
		t.Errorf("Verified create valid request should go through, errs=%v", errs)
		return
	}
	serverResponsePtr, ok = <-channel
	if !ok || serverResponsePtr.Result != Success {
		t.Errorf("Verified valid create request should succeed, result:%v", *serverResponsePtr)
		return
	}

	requestBytes = generateUserReadRequest("ISSUER", "CERTIFIER", []string{"USER"})
	channel, errs = MakeRequest(requestBytes)
	if len(errs) > 0 {
		t.Errorf("Read request with inexistent certifier should go through, errs=%v", errs)
		return
	}
	serverResponsePtr, ok = <-channel
	if !ok || serverResponsePtr.Result != Success {
		t.Errorf("Read request of existing should succeed, result:%v", *serverResponsePtr)
		return
	}
	if len(serverResponsePtr.Data) != 1 || !reflect.DeepEqual(*userObjectPtr, serverResponsePtr.Data[0]) {
		t.Errorf("Read request response doesn't match user expected.\nexpected=%v\nresult=%v", *userObjectPtr, serverResponsePtr.Data[0])
		return
	}

	ShutdownServer()
}

/*
	Create requests
*/

func TestUnknownIssuerCreateRequest(t *testing.T) {
	initErr := resetAndStartServer(multipleWorkersConfig())
	if initErr != nil {
		t.Errorf(initErr.Error())
		return
	}

	requestBytes, _, _, _ := generateUserCreateRequest("ISSUER", "CERTIFIER", "USER", true)

	channel, errs := MakeRequest(requestBytes)
	if len(errs) > 0 {
		t.Errorf("Create request with inexistent issuer should go through, errs=%v", errs)
		return
	}
	serverResponsePtr, ok := <-channel
	if !ok || serverResponsePtr.Result != IssuerUnknownError {
		t.Errorf("Create request with inexistent user should fail, result:%v", *serverResponsePtr)
		return
	}

	ShutdownServer()
}

func TestUnknownCertifierCreateRequest(t *testing.T) {
	initErr := resetAndStartServer(multipleWorkersConfig())
	if initErr != nil {
		t.Errorf(initErr.Error())
		return
	}

	requestBytes, _, _, _ := generateUserCreateRequest("ADMIN", "ADMIN", "ADMIN", true)
	channel, errs := MakeUnverifiedRequest(requestBytes)
	if len(errs) != 0 {
		t.Errorf("Unverified create request with inexistent issuer/certifier should go through, errs=%v", errs)
		return
	}
	serverResponsePtr, ok := <-channel
	if !ok || serverResponsePtr.Result != Success {
		t.Errorf("Unverified create request with inexistent issuer/certifier should succeed, result:%v", *serverResponsePtr)
		return
	}

	requestBytes, _, _, _ = generateUserCreateRequest("ADMIN", "CERTIFIER", "USER", true)
	channel, errs = MakeRequest(requestBytes)
	if len(errs) > 0 {
		t.Errorf("Create request with inexistent certifier should go through, errs=%v", errs)
		return
	}
	serverResponsePtr, ok = <-channel
	if !ok || serverResponsePtr.Result != CertifierUnknownError {
		t.Errorf("Create request with inexistent certifier should fail, result:%v", *serverResponsePtr)
		return
	}

	ShutdownServer()
}

func TestMissingPermissionVerifiedCreateRequest(t *testing.T) {
	initErr := resetAndStartServer(multipleWorkersConfig())
	if initErr != nil {
		t.Errorf(initErr.Error())
		return
	}

	requestBytes, _, _, _ := generateUserCreateRequest("ISSUER", "ISSUER", "ISSUER", true)
	channel, errs := MakeUnverifiedRequest(requestBytes)
	if len(errs) != 0 {
		t.Errorf("Unverified create request with inexistent issuer/certifier should go through, errs=%v", errs)
		return
	}
	serverResponsePtr, ok := <-channel
	if !ok || serverResponsePtr.Result != Success {
		t.Errorf("Unverified create request with inexistent issuer/certifier should succeed, result:%v", *serverResponsePtr)
		return
	}

	requestBytes, _, _, _ = generateUserCreateRequest("CERTIFIER", "CERTIFIER", "CERTIFIER", false)
	channel, errs = MakeUnverifiedRequest(requestBytes)
	if len(errs) != 0 {
		t.Errorf("Unverified create request with inexistent issuer/certifier should go through, errs=%v", errs)
		return
	}
	serverResponsePtr, ok = <-channel
	if !ok || serverResponsePtr.Result != Success {
		t.Errorf("Unverified create request with inexistent issuer/certifier should succeed, result:%v", *serverResponsePtr)
		return
	}

	requestBytes, _, _, _ = generateUserCreateRequest("ISSUER", "CERTIFIER", "USER", true)
	channel, errs = MakeRequest(requestBytes)
	if len(errs) > 0 {
		t.Errorf("Create request with inexistent certifier should go through, errs=%v", errs)
		return
	}
	serverResponsePtr, ok = <-channel
	if !ok || serverResponsePtr.Result != CertifierPermissionsError {
		t.Errorf("Create request with issuer/certifier should fail because of permissions, result:%v", *serverResponsePtr)
		return
	}

	ShutdownServer()
}

func TestSuccessfulVerifiedCreateRequest(t *testing.T) {
	initErr := resetAndStartServer(multipleWorkersConfig())
	if initErr != nil {
		t.Errorf(initErr.Error())
		return
	}

	requestBytes, _, _, _ := generateUserCreateRequest("ISSUER", "ISSUER", "ISSUER", true)
	channel, errs := MakeUnverifiedRequest(requestBytes)
	if len(errs) != 0 {
		t.Errorf("Unverified create request with inexistent issuer/certifier should go through, errs=%v", errs)
		return
	}
	serverResponsePtr, ok := <-channel
	if !ok || serverResponsePtr.Result != Success {
		t.Errorf("Unverified create request with inexistent issuer/certifier should succeed, result:%v", *serverResponsePtr)
		return
	}

	requestBytes, _, _, _ = generateUserCreateRequest("CERTIFIER", "CERTIFIER", "CERTIFIER", true)
	channel, errs = MakeUnverifiedRequest(requestBytes)
	if len(errs) != 0 {
		t.Errorf("Unverified create request with inexistent issuer/certifier should go through, errs=%v", errs)
		return
	}
	serverResponsePtr, ok = <-channel
	if !ok || serverResponsePtr.Result != Success {
		t.Errorf("Unverified create request with inexistent issuer/certifier should succeed, result:%v", *serverResponsePtr)
		return
	}

	requestBytes, userObjectPtr, _, _ := generateUserCreateRequest("ISSUER", "CERTIFIER", "USER", true)
	channel, errs = MakeRequest(requestBytes)
	if len(errs) > 0 {
		t.Errorf("Create request with inexistent certifier should go through, errs=%v", errs)
		return
	}
	serverResponsePtr, ok = <-channel
	if !ok || serverResponsePtr.Result != Success {
		t.Errorf("Create request with issuer/certifier should succeed, result:%v", *serverResponsePtr)
		return
	}
	if len(serverResponsePtr.Data) != 1 || !reflect.DeepEqual(*userObjectPtr, serverResponsePtr.Data[0]) {
		t.Errorf("Create request response doesn't match user expected.\nexpected=%v\nresult=%v", *userObjectPtr, serverResponsePtr.Data[0])
		return
	}

	ShutdownServer()
}
