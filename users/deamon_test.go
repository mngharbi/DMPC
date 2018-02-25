package users

import (
	"reflect"
	"testing"
)

/*
	General tests
*/

func TestStartShutdownSingleWorker(t *testing.T) {
	if !resetAndStartServer(t, singleWorkerConfig()) {
		return
	}
	ShutdownServer()
}

func TestStartShutdown(t *testing.T) {
	if !resetAndStartServer(t, multipleWorkersConfig()) {
		return
	}
	ShutdownServer()
}

func TestMalformattedRequest(t *testing.T) {
	if !resetAndStartServer(t, multipleWorkersConfig()) {
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
	if !resetAndStartServer(t, multipleWorkersConfig()) {
		return
	}

	_, errs := makeUserReadRequest("ISSUER", "CERTIFIER", []string{})
	if len(errs) == 0 {
		t.Error("Read request with no users should fail")
	}

	ShutdownServer()
}

func TestUnknownIssuerReadRequest(t *testing.T) {
	if !resetAndStartServer(t, multipleWorkersConfig()) {
		return
	}

	serverResponsePtr, ok, success := makeAndGetUserReadRequest(t, "ISSUER", "CERTIFIER", []string{"USER"})
	if !success {
		return
	}
	if !ok || serverResponsePtr.Result != IssuerUnknownError {
		t.Errorf("Read request with inexistent issuer should fail, result:%v", *serverResponsePtr)
		return
	}

	ShutdownServer()
}

func TestUnknownCertifierReadRequest(t *testing.T) {
	if !resetAndStartServer(t, multipleWorkersConfig()) {
		return
	}

	// Create issuer
	if !createUnverifiedUser(t, "ISSUER", false, true, false, false, false, false) {
		return
	}

	// Make read request
	serverResponsePtr, ok, success := makeAndGetUserReadRequest(t, "ISSUER", "CERTIFIER", []string{"USER"})
	if !success {
		return
	}
	if !ok || serverResponsePtr.Result != CertifierUnknownError {
		t.Errorf("Read request with inexistent certifier should fail, result:%v", *serverResponsePtr)
		return
	}

	ShutdownServer()
}

func TestUnknownSubjectReadRequest(t *testing.T) {
	if !resetAndStartServer(t, multipleWorkersConfig()) {
		return
	}

	// Create issuer and certifier
	if !createIssuerAndCertifier(t,
		false, true, false, false, false, false,
		false, true, false, false, false, false,
	) {
		return
	}

	// Make read request
	serverResponsePtr, ok, success := makeAndGetUserReadRequest(t, "ISSUER", "CERTIFIER", []string{"USER"})
	if !success {
		return
	}
	if !ok || serverResponsePtr.Result != SubjectUnknownError {
		t.Errorf("Read request with inexistent subject should fail, result:%v", *serverResponsePtr)
		return
	}

	ShutdownServer()
}

func TestExistentUserReadRequest(t *testing.T) {
	if !resetAndStartServer(t, multipleWorkersConfig()) {
		return
	}

	// Create issuer and certifier
	if !createIssuerAndCertifier(t,
		false, true, false, false, false, false,
		false, true, false, false, false, false,
	) {
		return
	}

	// Create user
	userObjectPtr, success := createUser(
		t, false, "ISSUER", "CERTIFIER", "USER", false, true, false, false, false, false,
	)
	if !success {
		return
	}

	// Make read request
	serverResponsePtr, ok, success := makeAndGetUserReadRequest(t, "ISSUER", "CERTIFIER", []string{"USER"})
	if !success {
		return
	}
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
	if !resetAndStartServer(t, multipleWorkersConfig()) {
		return
	}

	serverResponsePtr, ok, _, success := makeAndGetUserCreationRequest(
		t, false, "ISSUER", "CERTIFIER", "USER", false, true, false, false, false, false,
	)
	if !success {
		return
	}
	if !ok || serverResponsePtr.Result != IssuerUnknownError {
		t.Errorf("Create request with inexistent user should fail, result:%v", *serverResponsePtr)
		return
	}

	ShutdownServer()
}

func TestUnknownCertifierCreateRequest(t *testing.T) {
	if !resetAndStartServer(t, multipleWorkersConfig()) {
		return
	}

	// Create issuer
	if !createUnverifiedUser(t, "ISSUER", false, true, false, false, false, false) {
		return
	}

	serverResponsePtr, ok, _, success := makeAndGetUserCreationRequest(
		t, false, "ISSUER", "CERTIFIER", "USER", false, true, false, false, false, false,
	)
	if !success {
		return
	}
	if !ok || serverResponsePtr.Result != CertifierUnknownError {
		t.Errorf("Create request with inexistent certifier should fail, result:%v", *serverResponsePtr)
		return
	}

	ShutdownServer()
}

func TestMissingPermissionVerifiedCreateRequest(t *testing.T) {
	if !resetAndStartServer(t, multipleWorkersConfig()) {
		return
	}

	// Create issuer and certifier
	if !createIssuerAndCertifier(t,
		false, true, false, false, false, false,
		false, false, false, false, false, false,
	) {
		return
	}

	serverResponsePtr, ok, _, success := makeAndGetUserCreationRequest(
		t, false, "ISSUER", "CERTIFIER", "USER", false, true, false, false, false, false,
	)
	if !success {
		return
	}
	if !ok || serverResponsePtr.Result != CertifierPermissionsError {
		t.Errorf("Create request with issuer/certifier should fail because of permissions, result:%v", *serverResponsePtr)
		return
	}

	ShutdownServer()
}

func TestSuccessfulVerifiedCreateRequest(t *testing.T) {
	if !resetAndStartServer(t, multipleWorkersConfig()) {
		return
	}

	// Create issuer and certifier
	if !createIssuerAndCertifier(t,
		false, true, false, false, false, false,
		false, true, false, false, false, false,
	) {
		return
	}

	serverResponsePtr, ok, userObjectPtr, success := makeAndGetUserCreationRequest(
		t, false, "ISSUER", "CERTIFIER", "USER", false, true, false, false, false, false,
	)
	if !success {
		return
	}
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

/*
	Update requests
*/
func TestIdUpdateRequest(t *testing.T) {
	if !resetAndStartServer(t, multipleWorkersConfig()) {
		return
	}

	// Create issuer and certifier with all permissions
	if !createIssuerAndCertifier(t,
		true, true, true, true, true, true,
		true, true, true, true, true, true,
	) {
		return
	}
	// Create user
	_, success := createUser(
		t, false, "ISSUER", "CERTIFIER", "USER", false, false, false, false, false, false,
	)
	if !success {
		return
	}

	// Try to update id
	idStr := "userId"
	_, errs := makeUserUpdateRequest(
		"ISSUER", "CERTIFIER", []string{"id"}, &idStr, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	if len(errs) == 0 {
		t.Errorf("Update request for id should be ignored.")
		return
	}

	ShutdownServer()
}
