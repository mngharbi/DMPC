/*
	Test helpers
*/

package executor

import (
	"testing"
)

/*
	Server
*/

func resetAndStartServer(
	t *testing.T,
	conf Config,
	usersRequester UsersRequester,
	usersRequesterUnverified UsersRequester,
	responseReporter ResponseReporter,
	ticketGenerator TicketGenerator,
) bool {
	serverSingleton = server{}
	InitializeServer(usersRequester, usersRequesterUnverified, responseReporter, ticketGenerator, log, shutdownProgram)
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
