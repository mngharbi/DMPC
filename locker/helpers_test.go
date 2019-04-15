/*
	Test helpers
*/

package locker

import (
	"github.com/mngharbi/DMPC/core"
	"testing"
)

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

/*
	Request building
*/

func addLockingNeed(request *LockerRequest, lockType core.LockType, id string) {
	request.Needs = append(request.Needs, core.LockNeed{
		LockType: lockType,
		Id:       id,
	})
}
