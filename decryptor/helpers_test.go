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
	usersSignKeyRequester UsersSignKeyRequester,
	keyRequester KeyRequester,
	executorRequester ExecutorRequester,
) bool {
	serverSingleton = server{}
	InitializeServer(globalKey, usersSignKeyRequester, keyRequester, executorRequester, log, shutdownProgram)
	err := StartServer(conf)
	if err != nil {
		t.Errorf(err.Error())
		return false
	}
	return true
}

func makeRequestAndGetResult(
	t *testing.T,
	requestBytes []byte,
) (*DecryptorResponse, bool) {
	channel, errs := MakeRequest(requestBytes)
	if len(errs) != 0 {
		t.Errorf("Decryptor should pass along request.")
		return nil, false
	}
	nativeRespPtr := <-channel
	return (*nativeRespPtr).(*DecryptorResponse), true
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
