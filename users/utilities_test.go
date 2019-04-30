package users

import (
	"strconv"
	"testing"
)

func makeUsers(t *testing.T) bool {
	if !resetAndStartServer(t, multipleWorkersConfig()) {
		return false
	}

	// Create issuer and certifier
	if !createIssuerAndCertifier(t,
		false, false, true, false, false, false, false,
		false, false, true, false, false, false, false,
	) {
		return false
	}

	// Create 3 users
	for i := 0; i < 3; i++ {
		userSuffix := "_" + strconv.Itoa(i)
		_, success := createUser(
			t, false, "ISSUER", "CERTIFIER", "USER"+userSuffix, false, false, true, false, false, false, false,
		)
		if !success {
			return false
		}
	}

	return true
}

func TestGetSigningKeysById(t *testing.T) {
	if !makeUsers(t) {
		return
	}
	defer ShutdownServer()

	// Make valid signing keys read
	keys, err := GetSigningKeysById([]string{"USER_0", "USER_1", "USER_2"})
	if err != nil || len(keys) != 3 {
		t.Errorf("Getting signing keys failed. err=%+v", err)
	}

	// Request one inexistent id
	keys, err = GetSigningKeysById([]string{"USER_0", "USER_1", "USER_4"})
	if err == nil {
		t.Errorf("Getting inexistent signing keys should fail. keys=%+v", keys)
	}

	// Request no ids
	keys, err = GetSigningKeysById([]string{})
	if err == nil {
		t.Errorf("Getting signing keys without ids should fail. keys=%+v", keys)
	}
}

func TestGetChannelPermissionsByIds(t *testing.T) {
	if !makeUsers(t) {
		return
	}
	defer ShutdownServer()

	// Make valid signing keys read
	perms, err := GetChannelPermissionsByIds([]string{"USER_0", "USER_1", "USER_2"})
	if err != nil || len(perms) != 3 {
		t.Errorf("Getting channel permissions failed. err=%+v", err)
	}

	// Request one inexistent id
	perms, err = GetChannelPermissionsByIds([]string{"USER_0", "USER_1", "USER_4"})
	if err == nil {
		t.Errorf("Getting inexistent channel permissions should fail. perms=%+v", perms)
	}

	// Request no ids
	perms, err = GetChannelPermissionsByIds([]string{})
	if err == nil {
		t.Errorf("Getting channel permissions without ids should fail. perms=%+v", perms)
	}
}

func TestReadUserRecordsByIds(t *testing.T) {
	if !makeUsers(t) {
		return
	}
	defer ShutdownServer()

	// Make sure we attempt to read all user ids
	userRecords := readUserRecordsByIds(serverSingleton.store, []string{"USER_0", "USER_1", "USER_2", "NOT_USER_ID"})
	if len(userRecords) != 4 {
		t.Errorf("Reading user records failed. userRecords=%+v", userRecords)
	}

	// Existing users should not be nil
	if userRecords[0] == nil || userRecords[0].Id != "USER_0" ||
		userRecords[1] == nil || userRecords[1].Id != "USER_1" ||
		userRecords[2] == nil || userRecords[2].Id != "USER_2" {
		t.Errorf("Reading existing user records should return corresponding records. userRecords=%+v", userRecords)
	}

	// Non-existing users should be nil
	if userRecords[3] != nil {
		t.Errorf("Reading non-existing users should be nil . userRecords=%+v", userRecords)
	}
}
