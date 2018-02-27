package users

import (
	"reflect"
	"strings"
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
		"ISSUER", "CERTIFIER", []string{"id"}, getJanuaryDate(30), &idStr, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	if len(errs) == 0 {
		t.Errorf("Update request for id should be ignored.")
		return
	}

	ShutdownServer()
}

func TestEncKeyUpdateRequest(t *testing.T) {
	if !resetAndStartServer(t, multipleWorkersConfig()) {
		return
	}

	// Create issuer and certifier with no enc key update permissions
	if !createIssuerAndCertifier(t,
		true, true, true, true, true, true,
		true, true, true, false, true, true,
	) {
		return
	}
	// Create user
	userid := "USER"
	originalUserObjectPtr, success := createUser(
		t, false, "ISSUER", "CERTIFIER", userid, false, false, false, false, false, false,
	)
	if !success {
		return
	}

	// Try to update encKey
	publicKey := generatePublicKey()
	encKeyString := pemEncodeKey(publicKey)
	encKeyStringJson := jsonPemEncodeKey(publicKey)
	encKeyStringJson = strings.TrimSuffix(encKeyStringJson, `"`)
	encKeyStringJson = strings.TrimPrefix(encKeyStringJson, `"`)
	// Without subject id
	serverResponsePtr, ok, success := makeAndGetUserUpdateRequest(
		t, "ISSUER", "CERTIFIER", []string{"encKey"}, getJanuaryDate(1), nil, &encKeyStringJson, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	if !success {
		return
	}
	if !ok || serverResponsePtr.Result != SubjectUnknownError {
		t.Errorf("Update request to encKey without subject id should fail, result:%v", *serverResponsePtr)
		return
	}
	// With subject id
	serverResponsePtr, ok, success = makeAndGetUserUpdateRequest(
		t, "ISSUER", "CERTIFIER", []string{"encKey"}, getJanuaryDate(1), &userid, &encKeyStringJson, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	if !success {
		return
	}
	if !ok || serverResponsePtr.Result != CertifierPermissionsError {
		t.Errorf("Update request to encKey without permissions should fail, result:%v", *serverResponsePtr)
		return
	}

	// Create certifier with only enc key update permissions and use it to update again
	_, success = createUser(
		t, false, "ISSUER", "CERTIFIER", "ENCKEY_CERTIFIER", false, false, false, true, false, false,
	)
	if !success {
		return
	}

	// Try with stale date
	serverResponsePtr, ok, success = makeAndGetUserUpdateRequest(
		t, "ISSUER", "ENCKEY_CERTIFIER", []string{"encKey"}, getJanuaryDate(1), &userid, &encKeyStringJson, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	if !success {
		return
	}
	if !ok || serverResponsePtr.Result != Success {
		t.Errorf("Update request to encKey with permissions should succeed, result:%v", *serverResponsePtr)
		return
	}
	// Expect no changes
	if len(serverResponsePtr.Data) != 1 || !reflect.DeepEqual(*originalUserObjectPtr, serverResponsePtr.Data[0]) {
		t.Errorf("Stale encKey update should succeed but not affect anything.\n expected=%+v\n result=%+v", *originalUserObjectPtr, serverResponsePtr.Data[0])
	}

	// Try with recent date
	serverResponsePtr, ok, success = makeAndGetUserUpdateRequest(
		t, "ISSUER", "ENCKEY_CERTIFIER", []string{"encKey"}, getJanuaryDate(30), &userid, &encKeyStringJson, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	if !success {
		return
	}
	if !ok || serverResponsePtr.Result != Success {
		t.Errorf("Update request to encKey with permissions should succeed, result:%v", *serverResponsePtr)
		return
	}
	// Expect changes to enc key and updated at
	expectedAfterUpdates := *originalUserObjectPtr
	expectedAfterUpdates.EncKey = encKeyString
	expectedAfterUpdates.UpdatedAt = getJanuaryDate(30)
	if len(serverResponsePtr.Data) != 1 || !reflect.DeepEqual(expectedAfterUpdates, serverResponsePtr.Data[0]) {
		t.Errorf("Recent encKey update should succeed but and affect key and timestamps.\n expected=%+v\n result=%+v", expectedAfterUpdates, serverResponsePtr.Data[0])
	}

	ShutdownServer()
}

func TestSignKeyUpdateRequest(t *testing.T) {
	if !resetAndStartServer(t, multipleWorkersConfig()) {
		return
	}

	// Create issuer and certifier with no sign key update permissions
	if !createIssuerAndCertifier(t,
		true, true, true, true, true, true,
		true, true, true, true, false, true,
	) {
		return
	}
	// Create user
	userid := "USER"
	originalUserObjectPtr, success := createUser(
		t, false, "ISSUER", "CERTIFIER", userid, false, false, false, false, false, false,
	)
	if !success {
		return
	}

	// Try to update signKey
	publicKey := generatePublicKey()
	signKeyString := pemEncodeKey(publicKey)
	signKeyStringJson := jsonPemEncodeKey(publicKey)
	signKeyStringJson = strings.TrimSuffix(signKeyStringJson, `"`)
	signKeyStringJson = strings.TrimPrefix(signKeyStringJson, `"`)
	// Without subject id
	serverResponsePtr, ok, success := makeAndGetUserUpdateRequest(
		t, "ISSUER", "CERTIFIER", []string{"signKey"}, getJanuaryDate(1), nil, nil, &signKeyStringJson, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	if !success {
		return
	}
	if !ok || serverResponsePtr.Result != SubjectUnknownError {
		t.Errorf("Update request to signKey without subject id should fail, result:%v", *serverResponsePtr)
		return
	}
	// With subject id
	serverResponsePtr, ok, success = makeAndGetUserUpdateRequest(
		t, "ISSUER", "CERTIFIER", []string{"signKey"}, getJanuaryDate(1), &userid, nil, &signKeyStringJson, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	if !success {
		return
	}
	if !ok || serverResponsePtr.Result != CertifierPermissionsError {
		t.Errorf("Update request to signKey without permissions should fail, result:%v", *serverResponsePtr)
		return
	}

	// Create certifier with only sign key update permissions and use it to update again
	_, success = createUser(
		t, false, "ISSUER", "CERTIFIER", "SIGNKEY_CERTIFIER", false, false, false, false, true, false,
	)
	if !success {
		return
	}

	// Try with stale date
	serverResponsePtr, ok, success = makeAndGetUserUpdateRequest(
		t, "ISSUER", "SIGNKEY_CERTIFIER", []string{"signKey"}, getJanuaryDate(1), &userid, nil, &signKeyStringJson, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	if !success {
		return
	}
	if !ok || serverResponsePtr.Result != Success {
		t.Errorf("Update request to signKey with permissions should succeed, result:%v", *serverResponsePtr)
		return
	}
	// Expect no changes
	if len(serverResponsePtr.Data) != 1 || !reflect.DeepEqual(*originalUserObjectPtr, serverResponsePtr.Data[0]) {
		t.Errorf("Stale signKey update should succeed but not affect anything.\n expected=%+v\n result=%+v", *originalUserObjectPtr, serverResponsePtr.Data[0])
	}

	// Try with recent date
	serverResponsePtr, ok, success = makeAndGetUserUpdateRequest(
		t, "ISSUER", "SIGNKEY_CERTIFIER", []string{"signKey"}, getJanuaryDate(30), &userid, nil, &signKeyStringJson, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
	)
	if !success {
		return
	}
	if !ok || serverResponsePtr.Result != Success {
		t.Errorf("Update request to signKey with permissions should succeed, result:%v", *serverResponsePtr)
		return
	}
	// Expect changes to sign key and updated at
	expectedAfterUpdates := *originalUserObjectPtr
	expectedAfterUpdates.SignKey = signKeyString
	expectedAfterUpdates.UpdatedAt = getJanuaryDate(30)
	if len(serverResponsePtr.Data) != 1 || !reflect.DeepEqual(expectedAfterUpdates, serverResponsePtr.Data[0]) {
		t.Errorf("Recent signKey update should succeed but and affect key and timestamps.\n expected=%+v\n result=%+v", expectedAfterUpdates, serverResponsePtr.Data[0])
	}

	ShutdownServer()
}

func TestPermissionsUpdateRequest(t *testing.T) {
	if !resetAndStartServer(t, multipleWorkersConfig()) {
		return
	}

	// Create issuer and certifier with no permission update permissions
	if !createIssuerAndCertifier(t,
		true, true, true, true, true, true,
		true, true, true, true, true, false,
	) {
		return
	}

	// All permission fields possible
	// @TODO: No hardcoding
	permissionFields := []string{
		"permissions.channel.add",
		"permissions.user.add",
		"permissions.user.remove",
		"permissions.user.encKeyUpdate",
		"permissions.user.signKeyUpdate",
		"permissions.user.permissionsUpdate",
	}
	for permissionIndex, permissionType := range permissionFields {
		// Create user
		userid := "USER"
		originalUserObjectPtr, success := createUser(
			t, false, "ISSUER", "CERTIFIER", userid, false, false, false, false, false, false,
		)
		if !success {
			return
		}

		// Make an array of arguments to pass to change function
		permissionsChanges := []*bool{nil, nil, nil, nil, nil, nil}
		changedPermPtr := true
		permissionsChanges[permissionIndex] = &changedPermPtr

		// Without subject id
		serverResponsePtr, ok, success := makeAndGetUserUpdateRequest(
			t, "ISSUER", "CERTIFIER", []string{permissionType}, getJanuaryDate(1), nil, nil, nil, permissionsChanges[0], permissionsChanges[1], permissionsChanges[2], permissionsChanges[3], permissionsChanges[4], permissionsChanges[5], nil, nil, nil, nil,
		)
		if !success {
			return
		}
		if !ok || serverResponsePtr.Result != SubjectUnknownError {
			t.Errorf("Update request to permission %v without subject id should fail, result:%v", permissionType, *serverResponsePtr)
			return
		}
		// With subject id
		serverResponsePtr, ok, success = makeAndGetUserUpdateRequest(
			t, "ISSUER", "CERTIFIER", []string{permissionType}, getJanuaryDate(1), &userid, nil, nil, permissionsChanges[0], permissionsChanges[1], permissionsChanges[2], permissionsChanges[3], permissionsChanges[4], permissionsChanges[5], nil, nil, nil, nil,
		)
		if !success {
			return
		}
		if !ok || serverResponsePtr.Result != CertifierPermissionsError {
			t.Errorf("Update request to permission %v without update permissions should fail, result:%v", permissionType, *serverResponsePtr)
			return
		}

		// Create certifier with only update permissions permission and use it to update again
		_, success = createUser(
			t, false, "ISSUER", "CERTIFIER", permissionType+"_CERTIFIER", false, false, false, false, false, true,
		)
		if !success {
			return
		}

		// Try with stale date
		serverResponsePtr, ok, success = makeAndGetUserUpdateRequest(
			t, "ISSUER", permissionType+"_CERTIFIER", []string{permissionType}, getJanuaryDate(1), &userid, nil, nil, permissionsChanges[0], permissionsChanges[1], permissionsChanges[2], permissionsChanges[3], permissionsChanges[4], permissionsChanges[5], nil, nil, nil, nil,
		)
		if !success {
			return
		}
		if !ok || serverResponsePtr.Result != Success {
			t.Errorf("Update request to permission %v with permissions should succeed, result:%v", permissionType, *serverResponsePtr)
			return
		}
		// Expect no changes
		if len(serverResponsePtr.Data) != 1 || !reflect.DeepEqual(*originalUserObjectPtr, serverResponsePtr.Data[0]) {
			t.Errorf("Stale encKey update should succeed but not affect anything.\n expected=%+v\n result=%+v", *originalUserObjectPtr, serverResponsePtr.Data[0])
		}

		// Try with recent date
		serverResponsePtr, ok, success = makeAndGetUserUpdateRequest(
			t, "ISSUER", permissionType+"_CERTIFIER", []string{permissionType}, getJanuaryDate(30), &userid, nil, nil, permissionsChanges[0], permissionsChanges[1], permissionsChanges[2], permissionsChanges[3], permissionsChanges[4], permissionsChanges[5], nil, nil, nil, nil,
		)
		if !success {
			return
		}
		if !ok || serverResponsePtr.Result != Success {
			t.Errorf("Update request to permission %v with permissions should succeed, result:%v", permissionType, *serverResponsePtr)
			return
		}
		// Expect changes to enc key and updated at
		expectedAfterUpdates := *originalUserObjectPtr
		var expectedAfterUpdatesPermission *bool
		switch permissionType {
		case "permissions.channel.add":
			expectedAfterUpdatesPermission = &expectedAfterUpdates.Permissions.Channel.Add
		case "permissions.user.add":
			expectedAfterUpdatesPermission = &expectedAfterUpdates.Permissions.User.Add
		case "permissions.user.remove":
			expectedAfterUpdatesPermission = &expectedAfterUpdates.Permissions.User.Remove
		case "permissions.user.encKeyUpdate":
			expectedAfterUpdatesPermission = &expectedAfterUpdates.Permissions.User.EncKeyUpdate
		case "permissions.user.signKeyUpdate":
			expectedAfterUpdatesPermission = &expectedAfterUpdates.Permissions.User.SignKeyUpdate
		case "permissions.user.permissionsUpdate":
			expectedAfterUpdatesPermission = &expectedAfterUpdates.Permissions.User.PermissionsUpdate
		}
		*expectedAfterUpdatesPermission = true
		expectedAfterUpdates.UpdatedAt = getJanuaryDate(30)
		if len(serverResponsePtr.Data) != 1 || !reflect.DeepEqual(expectedAfterUpdates, serverResponsePtr.Data[0]) {
			t.Errorf("Recent permission %v update should succeed but and affect key and timestamps.\n expected=%+v\n result=%+v", permissionType, expectedAfterUpdates, serverResponsePtr.Data[0])
		}
	}

	ShutdownServer()
}

func TestDisableUpdateRequest(t *testing.T) {
	if !resetAndStartServer(t, multipleWorkersConfig()) {
		return
	}

	// Create issuer and certifier with no remove user permission
	if !createIssuerAndCertifier(t,
		true, true, true, true, true, true,
		true, true, false, true, true, true,
	) {
		return
	}
	// Create user
	userid := "USER"
	originalUserObjectPtr, success := createUser(
		t, false, "ISSUER", "CERTIFIER", userid, false, false, false, false, false, false,
	)
	if !success {
		return
	}

	// Try to update active boolean
	active := true
	// Without subject id
	serverResponsePtr, ok, success := makeAndGetUserUpdateRequest(
		t, "ISSUER", "CERTIFIER", []string{"active"}, getJanuaryDate(1), nil, nil, nil, nil, nil, nil, nil, nil, nil, &active, nil, nil, nil,
	)
	if !success {
		return
	}
	if !ok || serverResponsePtr.Result != SubjectUnknownError {
		t.Errorf("Update request to active without subject id should fail, result:%v", *serverResponsePtr)
		return
	}
	// With subject id
	serverResponsePtr, ok, success = makeAndGetUserUpdateRequest(
		t, "ISSUER", "CERTIFIER", []string{"active"}, getJanuaryDate(1), &userid, nil, nil, nil, nil, nil, nil, nil, nil, &active, nil, nil, nil,
	)
	if !success {
		return
	}
	if !ok || serverResponsePtr.Result != CertifierPermissionsError {
		t.Errorf("Update request to active without permissions should fail, result:%v", *serverResponsePtr)
		return
	}

	// Create certifier with only user remove update permissions and use it to update again
	_, success = createUser(
		t, false, "ISSUER", "CERTIFIER", "REMOVE_CERTIFIER", false, false, true, false, false, false,
	)
	if !success {
		return
	}

	// Try with stale date
	serverResponsePtr, ok, success = makeAndGetUserUpdateRequest(
		t, "ISSUER", "REMOVE_CERTIFIER", []string{"active"}, getJanuaryDate(1), &userid, nil, nil, nil, nil, nil, nil, nil, nil, &active, nil, nil, nil,
	)
	if !success {
		return
	}
	if !ok || serverResponsePtr.Result != Success {
		t.Errorf("Update request to active with permissions should succeed, result:%v", *serverResponsePtr)
		return
	}
	// Expect no changes
	if len(serverResponsePtr.Data) != 1 || !reflect.DeepEqual(*originalUserObjectPtr, serverResponsePtr.Data[0]) {
		t.Errorf("Stale active update should succeed but not affect anything.\n expected=%+v\n result=%+v", *originalUserObjectPtr, serverResponsePtr.Data[0])
	}

	// Try with recent date
	serverResponsePtr, ok, success = makeAndGetUserUpdateRequest(
		t, "ISSUER", "REMOVE_CERTIFIER", []string{"active"}, getJanuaryDate(30), &userid, nil, nil, nil, nil, nil, nil, nil, nil, &active, nil, nil, nil,
	)
	if !success {
		return
	}
	if !ok || serverResponsePtr.Result != Success {
		t.Errorf("Update request to active with permissions should succeed, result:%v", *serverResponsePtr)
		return
	}
	// Expect changes to sign key and updated at
	expectedAfterUpdates := *originalUserObjectPtr
	expectedAfterUpdates.Active = true
	expectedAfterUpdates.DisabledAt = getJanuaryDate(30)
	expectedAfterUpdates.UpdatedAt = getJanuaryDate(30)
	if len(serverResponsePtr.Data) != 1 || !reflect.DeepEqual(expectedAfterUpdates, serverResponsePtr.Data[0]) {
		t.Errorf("Recent active update should succeed but and affect key and timestamps.\n expected=%+v\n result=%+v", expectedAfterUpdates, serverResponsePtr.Data[0])
	}

	ShutdownServer()
}
