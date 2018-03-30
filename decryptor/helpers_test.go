/*
	Test helpers
*/

package decryptor

import (
	"crypto/rsa"
	"testing"
)

/*
	Server
*/

func resetAndStartServer(
	t *testing.T,
	conf Config,
	globalKey *rsa.PrivateKey,
	usersRequester UsersDecodedRequester,
	keyRequester KeyRequester,
	executorRequester ExecutorRequester,
) bool {
	serverSingleton = server{}
	InitializeServer(globalKey, usersRequester, keyRequester, executorRequester)
	err := StartServer(conf)
	if err != nil {
		t.Errorf(err.Error())
		return false
	}
	return true
}

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
