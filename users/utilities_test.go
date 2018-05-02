package users

import (
	"strconv"
	"testing"
)

func TestGetSigningKeysById(t *testing.T) {
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

	// Create 3 users
	for i := 0; i < 3; i++ {
		userSuffix := "_" + strconv.Itoa(i)
		_, success := createUser(
			t, false, "ISSUER", "CERTIFIER", "USER"+userSuffix, false, true, false, false, false, false,
		)
		if !success {
			return
		}
	}

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

	ShutdownServer()
}
