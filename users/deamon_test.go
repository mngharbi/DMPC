package users

import (
	"crypto/rsa"
	"encoding/json"
	"testing"
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

func generateUserCreateRequest(issuerId string, certifierId string, userId string) (request []byte, encKey *rsa.PublicKey, signKey *rsa.PublicKey) {
	encKey = generatePublicKey()
	encKeyStringEncoded := jsonPemEncodeKey(encKey)
	var encKeyStringDecoded string
	json.Unmarshal([]byte(encKeyStringEncoded), &encKeyStringDecoded)
	signKey = generatePublicKey()
	signKeyStringEncoded := jsonPemEncodeKey(signKey)
	var signKeyStringDecoded string
	json.Unmarshal([]byte(signKeyStringEncoded), &signKeyStringDecoded)

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
					"add": true,
					"remove": false,
					"encKeyUpdate": false,
					"signKeyUpdate": false,
					"permissionsUpdate": false
				}
			},
			"active": true
		},
		"timestamp": "2018-01-13T23:53:00Z"
	}`)

	return
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

	requestBytes := []byte(`{
		"type": 2,
		"issuerId": "ISSUER",
		"certifierId": "CERTIFIER",
		"fields": []
	}`)

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

	requestBytes := []byte(`{
		"type": 2,
		"issuerId": "ISSUER",
		"certifierId": "CERTIFIER",
		"fields": ["USER"]
	}`)

	channel, errs := MakeRequest(requestBytes)
	if len(errs) > 0 {
		t.Errorf("Read request with inexistent issuer should go through, errs=%v", errs)
		return
	}

	serverResponsePtr, ok := <-channel
	if !ok || serverResponsePtr.Result != IssuerUnknownError {
		t.Errorf("Read request with inexistent user shoud fail, result:%v", *serverResponsePtr)
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

	requestBytes, _, _ := generateUserCreateRequest("ISSUER", "CERTIFIER", "USER")

	channel, errs := MakeRequest(requestBytes)
	if len(errs) > 0 {
		t.Errorf("Create request with inexistent issuer should go through, errs=%v", errs)
		return
	}

	serverResponsePtr, ok := <-channel
	if !ok || serverResponsePtr.Result != IssuerUnknownError {
		t.Errorf("Create request with inexistent user shoud fail, result:%v", *serverResponsePtr)
		return
	}

	ShutdownServer()
}
